package config

import "testing"

func TestJWTSecretRequired(t *testing.T) {
	t.Chdir(t.TempDir())
	t.Setenv("DATABASE_URL", "postgres://localhost/books")
	t.Setenv("JWT_SECRET", "")
	if _, err := Load(); err == nil {
		t.Fatal("missing JWT secret should fail startup")
	}
	t.Setenv("JWT_SECRET", "mindstack")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWTSecret != "mindstack" {
		t.Fatal("JWT secret was not loaded from environment")
	}
}
