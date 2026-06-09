package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Profile struct {
	Type        string `json:"type"`
	Username    string `json:"user"`
	PasswordEnv string `json:"passwordEnv,omitempty"`
	Host        string `json:"host"`
	Port        string `json:"port"`
	Database    string `json:"dbname"`
}

type Config struct {
	DefaultProfile string             `json:"defaultProfile,omitempty"`
	Profiles       map[string]Profile `json:"profiles"`
}

func HomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".snapdb"), nil
}

func ConfigPath() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "config.json"), nil
}

func BackupsDir() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "backups"), nil
}

func MetadataPath() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "metadata.json"), nil
}

func SchedulesPath() (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "schedules.json"), nil
}

func CloudDir(provider string) (string, error) {
	home, err := HomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, "cloud", provider), nil
}

func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{Profiles: make(map[string]Profile)}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]Profile)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func EnsureDirs() error {
	home, err := HomeDir()
	if err != nil {
		return err
	}

	backups, err := BackupsDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(home, 0700); err != nil {
		return err
	}
	return os.MkdirAll(backups, 0700)
}
