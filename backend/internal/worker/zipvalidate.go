package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var zipNameRe = regexp.MustCompile(`^([a-zA-Z0-9_-]+)_v([0-9A-Za-z._-]+)_(\d{8})\.zip$`)

type ZipMeta struct {
	Slug       string
	Version    string
	Date       string
	FileName   string
	FileCount  int
	HasRuntime bool
	HasCommonW2 bool
	HasCommonW1 bool
	HasLoader  bool
	InfoPath   string
}

func ParseZipFileName(name string) (*ZipMeta, error) {
	m := zipNameRe.FindStringSubmatch(name)
	if m == nil {
		return nil, fmt.Errorf("zip filename must match {slug}_v{version}_{yyyyMMdd}.zip")
	}
	return &ZipMeta{
		Slug:     m[1],
		Version:  m[2],
		Date:     m[3],
		FileName: name,
	}, nil
}

func ValidateExtracted(root string, expectedSlug string) (*ZipMeta, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var zipFile string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".zip") {
			zipFile = e.Name()
			break
		}
	}
	if zipFile == "" {
		// already extracted layout
		return validateLayout(root, expectedSlug, "")
	}
	meta, err := ParseZipFileName(zipFile)
	if err != nil {
		return nil, err
	}
	if meta.Slug != expectedSlug {
		return nil, fmt.Errorf("slug mismatch: zip=%s expected=%s", meta.Slug, expectedSlug)
	}
	return validateLayout(root, expectedSlug, meta.Version)
}

func validateLayout(root, expectedSlug, version string) (*ZipMeta, error) {
	meta := &ZipMeta{Slug: expectedSlug, Version: version}
	levelsDir := filepath.Join(root, "levels", expectedSlug)
	infoPath := filepath.Join(levelsDir, "info_"+expectedSlug)
	if _, err := os.Stat(infoPath); err != nil {
		return nil, fmt.Errorf("missing required bundle: levels/%s/info_%s", expectedSlug, expectedSlug)
	}
	meta.InfoPath = infoPath

	sceneCount := 0
	entries, err := os.ReadDir(levelsDir)
	if err != nil {
		return nil, fmt.Errorf("missing levels/%s directory", expectedSlug)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "s_") {
			sceneCount++
		}
		if name == "runtime" {
			meta.HasRuntime = true
		}
		meta.FileCount++
	}
	if sceneCount == 0 {
		return nil, fmt.Errorf("no scene bundles (s_*) found in levels/%s", expectedSlug)
	}

	if _, err := os.Stat(filepath.Join(root, "commonW2")); err == nil {
		meta.HasCommonW2 = true
		meta.FileCount++
	}
	if _, err := os.Stat(filepath.Join(root, "commonW1")); err == nil {
		meta.HasCommonW1 = true
		meta.FileCount++
	}
	if _, err := os.Stat(filepath.Join(root, "OC2LevelRuntimeLoader.dll")); err == nil {
		meta.HasLoader = true
		meta.FileCount++
	}
	meta.FileCount++ // info bundle counted
	return meta, nil
}
