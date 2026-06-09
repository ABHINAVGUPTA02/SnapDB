package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := &Config{
		DefaultProfile: "local",
		Profiles: map[string]Profile{
			"local": {
				Type:        "mysql",
				Username:    "root",
				PasswordEnv: "SNAPDB_MYSQL_PASSWORD",
				Host:        "localhost",
				Port:        "3306",
				Database:    "appdb",
			},
		},
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config permissions = %v, want 0600", got)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	if loaded.DefaultProfile != "local" {
		t.Fatalf("DefaultProfile = %q, want local", loaded.DefaultProfile)
	}
	if loaded.Profiles["local"].PasswordEnv != "SNAPDB_MYSQL_PASSWORD" {
		t.Fatalf("PasswordEnv was not persisted")
	}
}

func TestEnsureDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(home, ".snapdb", "backups")); err != nil {
		t.Fatalf("backups directory was not created: %v", err)
	}
}
