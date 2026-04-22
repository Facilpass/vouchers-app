package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AdminUser       string
	AdminPassBcrypt string
	SessionSecret   string
	StorageRoot     string
	ListenAddr      string
	PublicBaseURL   string
	MaxUploadMB     int
	WebPQuality     int
}

const minSessionSecretLen = 32

func Load() (*Config, error) {
	c := &Config{
		AdminUser:       os.Getenv("ADMIN_USER"),
		AdminPassBcrypt: os.Getenv("ADMIN_PASS_BCRYPT"),
		SessionSecret:   os.Getenv("SESSION_SECRET"),
		StorageRoot:     envDefault("STORAGE_ROOT", "/srv/vouchers"),
		ListenAddr:      envDefault("LISTEN_ADDR", ":8080"),
		PublicBaseURL:   os.Getenv("PUBLIC_BASE_URL"),
		MaxUploadMB:     envInt("MAX_UPLOAD_MB", 5),
		WebPQuality:     envInt("WEBP_QUALITY", 85),
	}

	if c.AdminUser == "" {
		return nil, fmt.Errorf("ADMIN_USER obrigatório")
	}
	if c.AdminPassBcrypt == "" {
		return nil, fmt.Errorf("ADMIN_PASS_BCRYPT obrigatório")
	}
	if len(c.SessionSecret) < minSessionSecretLen {
		return nil, fmt.Errorf("SESSION_SECRET precisa ter pelo menos %d chars", minSessionSecretLen)
	}
	if c.PublicBaseURL == "" {
		return nil, fmt.Errorf("PUBLIC_BASE_URL obrigatório")
	}
	if c.WebPQuality < 1 || c.WebPQuality > 100 {
		return nil, fmt.Errorf("WEBP_QUALITY fora do range 1-100: %d", c.WebPQuality)
	}
	if c.MaxUploadMB < 1 {
		return nil, fmt.Errorf("MAX_UPLOAD_MB inválido: %d", c.MaxUploadMB)
	}

	return c, nil
}

func envDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
