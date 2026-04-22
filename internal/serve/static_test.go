package serve

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupStorage(t *testing.T) string {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "loungebrahma/copa/camarote"), 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "loungebrahma/copa/camarote/ok.webp"),
		[]byte("RIFF\x00\x00\x00\x00WEBP"), 0640); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestServeOK(t *testing.T) {
	dir := setupStorage(t)
	h := NewHandler(dir)
	req := httptest.NewRequest("GET", "/loungebrahma/copa/camarote/ok.webp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "image/webp" {
		t.Errorf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("nosniff header faltando")
	}
}

func TestServe404(t *testing.T) {
	dir := setupStorage(t)
	h := NewHandler(dir)
	req := httptest.NewRequest("GET", "/loungebrahma/missing/file.webp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestServeDirectoryIs404(t *testing.T) {
	dir := setupStorage(t)
	h := NewHandler(dir)
	req := httptest.NewRequest("GET", "/loungebrahma/copa/camarote/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 {
		t.Errorf("status = %d, want 404 (no autoindex)", rec.Code)
	}
}

func TestServeTraversalBlocked(t *testing.T) {
	dir := setupStorage(t)
	h := NewHandler(dir)
	req := httptest.NewRequest("GET", "/../../../etc/passwd", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code == 200 {
		t.Error("traversal deveria ser 404/400, não 200")
	}
}

func TestServeRejectNonWebP(t *testing.T) {
	dir := setupStorage(t)
	if err := os.WriteFile(filepath.Join(dir, "evil.php"), []byte("<?php"), 0640); err != nil {
		t.Fatal(err)
	}
	h := NewHandler(dir)
	req := httptest.NewRequest("GET", "/evil.php", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 404 && rec.Code != 403 {
		t.Errorf("arquivo .php deveria ser rejeitado, status = %d", rec.Code)
	}
}
