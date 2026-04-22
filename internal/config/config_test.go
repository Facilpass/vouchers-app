package config

import (
	"os"
	"testing"
)

func TestLoadRequiredVars(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASS_BCRYPT", "$2a$12$abc")
	os.Setenv("SESSION_SECRET", "secret-64-chars-long-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	os.Setenv("PUBLIC_BASE_URL", "https://vouchers.facilpass.com.br")
	defer os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() erro: %v", err)
	}
	if cfg.AdminUser != "admin" {
		t.Errorf("AdminUser = %q, want admin", cfg.AdminUser)
	}
	if cfg.MaxUploadMB != 5 {
		t.Errorf("MaxUploadMB default = %d, want 5", cfg.MaxUploadMB)
	}
	if cfg.WebPQuality != 85 {
		t.Errorf("WebPQuality default = %d, want 85", cfg.WebPQuality)
	}
	if cfg.StorageRoot != "/srv/vouchers" {
		t.Errorf("StorageRoot default = %q, want /srv/vouchers", cfg.StorageRoot)
	}
}

func TestLoadMissingRequired(t *testing.T) {
	os.Clearenv()
	_, err := Load()
	if err == nil {
		t.Fatal("Load() deveria falhar sem env vars required")
	}
}

func TestLoadShortSessionSecret(t *testing.T) {
	os.Setenv("ADMIN_USER", "admin")
	os.Setenv("ADMIN_PASS_BCRYPT", "$2a$12$abc")
	os.Setenv("SESSION_SECRET", "short")
	os.Setenv("PUBLIC_BASE_URL", "https://x")
	defer os.Clearenv()

	_, err := Load()
	if err == nil {
		t.Fatal("Load() deveria rejeitar SESSION_SECRET curto")
	}
}
