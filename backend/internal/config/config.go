package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	COS      COSConfig      `yaml:"cos"`
	Database DatabaseConfig `yaml:"database"`
	Upload   UploadConfig   `yaml:"upload"`
	Parser   ParserConfig   `yaml:"parser"`
	Bootstrap BootstrapConfig `yaml:"bootstrap"`
}

type ServerConfig struct {
	APIPort    int    `yaml:"api_port"`
	PublicPort int    `yaml:"public_port"`
	JWTSecret  string `yaml:"jwt_secret"`
	CookieName string `yaml:"cookie_name"`
}

type COSConfig struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	BaseURL   string `yaml:"base_url"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type UploadConfig struct {
	MaxZipMB int    `yaml:"max_zip_mb"`
	TempDir  string `yaml:"temp_dir"`
}

type ParserConfig struct {
	Python string `yaml:"python"`
	Script string `yaml:"script"`
}

type BootstrapConfig struct {
	SuperAdminUsername string `yaml:"super_admin_username"`
	SuperAdminPassword string `yaml:"super_admin_password"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.APIPort == 0 {
		cfg.Server.APIPort = 14556
	}
	if cfg.Server.PublicPort == 0 {
		cfg.Server.PublicPort = 14558
	}
	if cfg.Server.CookieName == "" {
		cfg.Server.CookieName = "oc2_manager_token"
	}
	if cfg.Upload.MaxZipMB == 0 {
		cfg.Upload.MaxZipMB = 512
	}
	if cfg.Upload.TempDir == "" {
		cfg.Upload.TempDir = "./data/tmp"
	}
	if cfg.Parser.Python == "" {
		cfg.Parser.Python = "python3"
	}
	if cfg.Parser.Script == "" {
		cfg.Parser.Script = "./tools/inspect_info_bundle.py"
	}
	return &cfg, nil
}

func (c *COSConfig) Enabled() bool {
	return c.SecretID != "" && c.SecretKey != "" && c.Bucket != ""
}
