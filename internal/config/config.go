package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	AppEnv      string
	AppBaseUrl  string
	DatabaseUrl string
	JWTSecret   string
	TokenTTL    time.Duration
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
	SMTPFrom    string
	UploadDir   string
}

func Load() *Config {
	_ = godotenv.Load()

	dbUser := requireEnv("DB_USER")
	dbPass := requireEnv("DB_PASSWORD")
	dbHost := requireEnv("DB_HOST")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := requireEnv("DB_NAME")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPass, dbHost, dbPort, dbName, dbSSLMode,
	)

	jwtSecret := requireEnv("JWT_SECRET")
	ttlHours, err := strconv.Atoi(getEnv("JWT_TTL_HOURS", "72"))
	if err != nil {
		ttlHours = 72
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		AppEnv:      getEnv("APP_ENV", "production"),
		AppBaseUrl:  getEnv("APP_BASE_URL", "http://localhost:5173"),
		DatabaseUrl: dbUrl,
		JWTSecret:   jwtSecret,
		TokenTTL:    time.Duration(ttlHours) * time.Hour,
		SMTPHost:    os.Getenv("SMTP_HOST"),
		SMTPPort:    getEnv("SMTP_PORT", "587"),
		SMTPUser:    os.Getenv("SMTP_USER"),
		SMTPPass:    os.Getenv("SMTP_PASS"),
		SMTPFrom:    getEnv("SMTP_FROM", "noreply@example.com"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Environment variable %s is required but not set", key)
	}
	return val
}
