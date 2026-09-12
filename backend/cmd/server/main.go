package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/config"
	embedstatic "github.com/gua248/Overcooked2-DiyLevel-Manager/internal/embed"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/handler"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/middleware"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/storage"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/worker"
)

// @title OC2 DIY Level Manager API
// @version 1.0
// @description Overcooked2 custom level set manager
// @BasePath /
func main() {
	configPath := flag.String("config", "../configs/config.yaml", "config file path")
	mode := flag.String("mode", "api", "api|public")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		// fallback to example for dev
		cfg, err = config.Load(strings.Replace(*configPath, "config.yaml", "config.example.yaml", 1))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("using example config")
	}

	db, err := repository.Open(cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}
	migrationPath := filepath.Join(filepath.Dir(*configPath), "..", "backend", "migrations", "001_init.sql")
	if _, err := os.Stat(migrationPath); err != nil {
		migrationPath = filepath.Join("migrations", "001_init.sql")
	}
	migrationSQL, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatal("read migration: ", err)
	}
	if err := repository.Migrate(db, string(migrationSQL)); err != nil {
		log.Fatal(err)
	}

	userRepo := repository.NewUserRepo(db)
	setRepo := repository.NewSetRepo(db)
	versionRepo := repository.NewVersionRepo(db)
	entryRepo := repository.NewEntryRepo(db)
	jobRepo := repository.NewJobRepo(db)

	localStorage := filepath.Join(filepath.Dir(cfg.Database.Path), "cos-local")
	cosStorage, err := storage.NewCOSStorage(&cfg.COS, db, localStorage)
	if err != nil {
		log.Fatal(err)
	}

	authSvc := service.NewAuthService(userRepo, &cfg.Server)
	if err := authSvc.BootstrapSuperAdmin(context.Background(), cfg.Bootstrap.SuperAdminUsername, cfg.Bootstrap.SuperAdminPassword); err != nil {
		log.Fatal(err)
	}

	userSvc := service.NewUserService(userRepo)
	setSvc := service.NewSetService(setRepo, versionRepo, entryRepo, cosStorage)
	inspector := worker.NewPythonInspector(&cfg.Parser)
	queue := worker.NewQueue(jobRepo, versionRepo, entryRepo, setRepo, cosStorage, inspector, cfg)
	uploadSvc := service.NewUploadService(setSvc, versionRepo, jobRepo, queue, cfg.Upload.TempDir, cfg.Upload.MaxZipMB)

	authH := handler.NewAuthHandler(authSvc, cfg.Server.CookieName)
	setsH := handler.NewSetsHandler(setSvc)
	uploadH := handler.NewUploadHandler(uploadSvc)
	adminH := handler.NewAdminHandler(setSvc, userSvc)

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:14557", "http://127.0.0.1:14557"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(middleware.Auth(authSvc, cfg.Server.CookieName))

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/api/v1/local-files/*", func(w http.ResponseWriter, req *http.Request) {
		key := strings.TrimPrefix(req.URL.Path, "/api/v1/local-files/")
		key, _ = url.PathUnescape(key)
		path, err := cosStorage.ServeLocal(key)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		http.ServeFile(w, req, path)
	})

	r.Get("/swagger/doc.json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(swaggerDoc))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/auth/login", authH.Login)
		api.With(middleware.RequireAuth).Get("/auth/me", authH.Me)
		api.With(middleware.RequireAuth).Post("/auth/change-password", authH.ChangePassword)

		api.Get("/sets", setsH.ListPublic)
		api.Get("/sets/{slug}", setsH.GetBySlug)
		api.Get("/sets/{slug}/latest/download", setsH.DownloadLatest)
		api.Get("/sets/{slug}/versions/{version}/download", setsH.DownloadVersion)
		api.Get("/sets/{slug}/bundle", setsH.BundleDownload)

		api.Route("/me", func(me chi.Router) {
			me.Use(middleware.RequireAuth)
			me.Get("/sets", setsH.ListMy)
			me.Post("/sets", setsH.CreateSet)
			me.Patch("/sets/{slug}", setsH.PatchSet)
			me.Post("/sets/{slug}/upload", uploadH.Upload)
			me.Get("/parse-jobs/{id}", uploadH.GetJob)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.RequireRoles("admin", "super_admin"))
			admin.Get("/sets", adminH.AdminListSets)
			admin.Patch("/sets/{id}/status", adminH.AdminPatchSetStatus)
		})

		api.Route("/super", func(super chi.Router) {
			super.Use(middleware.RequireRoles("super_admin"))
			super.Get("/users", adminH.ListUsers)
			super.Post("/users", adminH.CreateUser)
			super.Patch("/users/{id}", adminH.PatchUser)
			super.Get("/invalid-sets", adminH.ListInvalidSets)
			super.Post("/sets/{id}/invalidate", adminH.InvalidateSet)
		})
	})

	port := cfg.Server.APIPort
	if *mode == "public" {
		port = cfg.Server.PublicPort
		if embedstatic.Available() {
			static := embedstatic.Handler()
			r.Get("/*", static.ServeHTTP)
		}
	}

	_ = os.MkdirAll(cfg.Upload.TempDir, 0o755)
	addr := fmt.Sprintf(":%d", port)
	log.Printf("listening on %s (mode=%s)", addr, *mode)
	log.Fatal(http.ListenAndServe(addr, r))
}

const swaggerDoc = `{
  "swagger": "2.0",
  "info": {"title": "OC2 DIY Level Manager API", "version": "1.0"},
  "basePath": "/",
  "paths": {
    "/api/v1/sets": {"get": {"tags": ["public"], "summary": "List all published level sets"}},
    "/api/v1/sets/{slug}": {"get": {"tags": ["public"], "summary": "Get level set detail"}},
    "/api/v1/sets/{slug}/latest/download": {"get": {"tags": ["public"], "summary": "Download latest version"}},
    "/api/v1/sets/{slug}/versions/{version}/download": {"get": {"tags": ["public"], "summary": "Download specific version"}},
    "/api/v1/sets/{slug}/bundle": {"get": {"tags": ["public"], "summary": "Bundle download latest"}},
    "/api/v1/auth/login": {"post": {"tags": ["auth"], "summary": "Login"}},
    "/api/v1/auth/me": {"get": {"tags": ["auth"], "summary": "Current user"}},
    "/api/v1/auth/change-password": {"post": {"tags": ["auth"], "summary": "Change password"}},
    "/api/v1/me/sets": {"get": {"tags": ["author"], "summary": "List my sets"}, "post": {"tags": ["author"], "summary": "Create set"}},
    "/api/v1/me/sets/{slug}": {"patch": {"tags": ["author"], "summary": "Update set"}},
    "/api/v1/me/sets/{slug}/upload": {"post": {"tags": ["author"], "summary": "Upload zip"}},
    "/api/v1/me/parse-jobs/{id}": {"get": {"tags": ["author"], "summary": "Parse job progress"}},
    "/api/v1/admin/sets": {"get": {"tags": ["admin"], "summary": "Admin list sets"}},
    "/api/v1/admin/sets/{id}/status": {"patch": {"tags": ["admin"], "summary": "Update set status"}},
    "/api/v1/super/users": {"get": {"tags": ["super"], "summary": "List users"}, "post": {"tags": ["super"], "summary": "Create user"}},
    "/api/v1/super/users/{id}": {"patch": {"tags": ["super"], "summary": "Update user"}},
    "/api/v1/super/invalid-sets": {"get": {"tags": ["super"], "summary": "List invalid sets"}},
    "/api/v1/super/sets/{id}/invalidate": {"post": {"tags": ["super"], "summary": "Invalidate set"}}
  }
}`
