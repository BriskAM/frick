package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultModel(t *testing.T) {
	if DefaultModel != "gemma-4-31b-it" {
		t.Errorf("Expected DefaultModel to be gemma-4-31b-it, got %s", DefaultModel)
	}
}

func TestLoadSaveConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "frick-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath failed: %v", err)
	}
	expectedPath := filepath.Join(tempDir, ".config", "frick", "config.json")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Expected model %s, got %s", DefaultModel, cfg.Model)
	}
	if cfg.ApiKey != "" {
		t.Errorf("Expected empty ApiKey, got %s", cfg.ApiKey)
	}

	cfg.ApiKey = "test-key"
	cfg.Model = "test-model"
	cfg.SafetyOverrides = map[string]string{"rm": "DANGER"}
	err = SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat config file: %v", err)
	}
	mode := info.Mode().Perm()
	if mode != 0600 {
		t.Errorf("Expected file mode 0600, got %o", mode)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if loaded.ApiKey != "test-key" {
		t.Errorf("Expected ApiKey test-key, got %s", loaded.ApiKey)
	}
	if loaded.Model != "test-model" {
		t.Errorf("Expected Model test-model, got %s", loaded.Model)
	}
	if loaded.SafetyOverrides["rm"] != "DANGER" {
		t.Errorf("Expected SafetyOverrides[rm] to be DANGER, got %s", loaded.SafetyOverrides["rm"])
	}
}
