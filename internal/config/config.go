package config

import (
	"cmp"
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Port               string
	DBUrl              string
	JWTSecret          string
	Secure             bool
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
}

func Read() (Config, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET environment variable is required")
	}

	secure, _ := strconv.ParseBool(os.Getenv("SECURE"))

	return Config{
		Port:      cmp.Or(os.Getenv("MURMUR_PORT"), "8080"),
		DBUrl:     cmp.Or(os.Getenv("DB_URL"), "postgres://murmur:murmur@localhost:5432/murmur?sslmode=disable"),
		JWTSecret: jwtSecret,
		Secure:    secure,

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  cmp.Or(os.Getenv("GOOGLE_REDIRECT_URL"), "http://localhost:8080/api/auth/google/callback"),
		FrontendURL:        cmp.Or(os.Getenv("FRONTEND_URL"), "http://localhost:5173"),
	}, nil
}
