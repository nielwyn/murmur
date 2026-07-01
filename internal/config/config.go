package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".murmurconfig.json"

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	JWTSecret       string `json:"jwt_secret"`
	Secure          bool   `json:"secure"`
}

func Read() (Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c *Config) SetUser(name string) error {
	c.CurrentUserName = name
	return write(*c)
}

func write(cfg Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func getConfigFilePath() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func defaultConfig() Config {
	return Config{
		DBUrl:     "postgres://murmur:murmur@localhost:5432/murmur?sslmode=disable",
		JWTSecret: "murmur-dev-secret-change-me",
	}
}
