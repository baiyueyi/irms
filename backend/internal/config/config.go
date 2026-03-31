package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DBHost                  string
	DBPort                  string
	DBName                  string
	DBUser                  string
	DBPass                  string
	JWTSecret               string
	CredentialEncryptionKey string
	Addr                    string
}

func FromEnv() (Config, error) {
	cfg := Config{
		DBHost:                  os.Getenv("DB_HOST"),
		DBPort:                  os.Getenv("DB_PORT"),
		DBName:                  os.Getenv("DB_NAME"),
		DBUser:                  os.Getenv("DB_USER"),
		DBPass:                  os.Getenv("DB_PASSWORD"),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		CredentialEncryptionKey: os.Getenv("CREDENTIAL_ENCRYPTION_KEY"),
		Addr:                    os.Getenv("SERVER_ADDR"),
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBName == "" || cfg.DBUser == "" || cfg.DBPass == "" {
		return cfg, fmt.Errorf("missing required DB_* environment variables")
	}
	if !strings.EqualFold(cfg.DBName, "irms") {
		return cfg, fmt.Errorf("DB_NAME must be irms")
	}
	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("missing JWT_SECRET")
	}
	if cfg.CredentialEncryptionKey == "" {
		return cfg, fmt.Errorf("missing CREDENTIAL_ENCRYPTION_KEY")
	}
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	return cfg, nil
}
