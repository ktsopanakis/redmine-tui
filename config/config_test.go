package config

import (
"os"
"path/filepath"
"testing"
)

func TestGetConfigPath(t *testing.T) {
path, err := GetConfigPath()
if err != nil {
t.Fatalf("GetConfigPath() failed: %v", err)
}

if path == "" {
t.Error("GetConfigPath() returned empty path")
}

if !filepath.IsAbs(path) {
t.Errorf("GetConfigPath() returned relative path: %s", path)
}

if filepath.Base(path) != "config.yaml" {
t.Errorf("GetConfigPath() filename = %s, want config.yaml", filepath.Base(path))
}
}

func TestLoadNonExistentConfig(t *testing.T) {
// Save original env
originalHome := os.Getenv("HOME")
defer os.Setenv("HOME", originalHome)

// Set to temp directory that doesn't have config
tempDir := t.TempDir()
os.Setenv("HOME", tempDir)

err := Load()
if err == nil {
t.Error("Load() should fail with non-existent config")
}
}
