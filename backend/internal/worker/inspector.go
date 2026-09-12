package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/config"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
)

type BundleInspector interface {
	Inspect(ctx context.Context, bundlePath, outDir, slug string) (*domain.BundleManifest, error)
}

type PythonInspector struct {
	python string
	script string
}

func NewPythonInspector(cfg *config.ParserConfig) *PythonInspector {
	return &PythonInspector{python: cfg.Python, script: cfg.Script}
}

func (p *PythonInspector) Inspect(ctx context.Context, bundlePath, outDir, slug string) (*domain.BundleManifest, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, p.python, p.script,
		"--bundle", bundlePath,
		"--out-dir", outDir,
		"--slug", slug,
	)
	cmd.Dir = filepath.Dir(p.script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bundle inspect failed: %s: %w", string(out), err)
	}
	manifestPath := filepath.Join(outDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("manifest not found: %w (output: %s)", err, string(out))
	}
	var m domain.BundleManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
