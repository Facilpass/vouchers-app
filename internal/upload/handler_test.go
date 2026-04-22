package upload

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeMultipart(t *testing.T, pathField string, files map[string][]byte) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("path", pathField)
	_ = w.WriteField("csrf", "test-csrf-token")
	for name, data := range files {
		part, err := w.CreateFormFile("files", name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = io.Copy(part, bytes.NewReader(data))
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func TestHandlerUploadJPEG(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(Options{
		StorageRoot:   dir,
		MaxUploadMB:   5,
		WebPQuality:   85,
		PublicBaseURL: "https://vouchers.example.com",
		CSRFCheck:     func(string) bool { return true },
	})

	jpg, err := os.ReadFile("../../testdata/sample.jpg")
	if err != nil {
		t.Fatal(err)
	}
	body, ct := makeMultipart(t, "loungebrahma/copa-do-brasil/camarote", map[string][]byte{
		"VOUCHER_BARRA.jpeg": jpg,
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	expected := filepath.Join(dir, "loungebrahma/copa-do-brasil/camarote/VOUCHER_BARRA.webp")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("arquivo não encontrado em %s: %v", expected, err)
	}
	if !strings.Contains(rec.Body.String(), "VOUCHER_BARRA.webp") {
		t.Errorf("body não menciona filename final: %s", rec.Body.String())
	}
}

func TestHandlerRejectInvalidMIME(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(Options{
		StorageRoot: dir,
		MaxUploadMB: 5,
		WebPQuality: 85,
		CSRFCheck:   func(string) bool { return true },
	})
	body, ct := makeMultipart(t, "tenant/evento/cat", map[string][]byte{
		"evil.jpg": []byte("<?php echo 'pwn'; ?>"),
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", rec.Code)
	}
}

func TestHandlerTimestampOnCollision(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(Options{
		StorageRoot:   dir,
		MaxUploadMB:   5,
		WebPQuality:   85,
		PublicBaseURL: "https://x",
		CSRFCheck:     func(string) bool { return true },
	})
	jpg, err := os.ReadFile("../../testdata/sample.jpg")
	if err != nil {
		t.Fatal(err)
	}

	body1, ct1 := makeMultipart(t, "a/b/c", map[string][]byte{"file.jpg": jpg})
	req1 := httptest.NewRequest(http.MethodPost, "/admin/upload", body1)
	req1.Header.Set("Content-Type", ct1)
	h.ServeHTTP(httptest.NewRecorder(), req1)

	body2, ct2 := makeMultipart(t, "a/b/c", map[string][]byte{"file.jpg": jpg})
	req2 := httptest.NewRequest(http.MethodPost, "/admin/upload", body2)
	req2.Header.Set("Content-Type", ct2)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)

	entries, _ := os.ReadDir(filepath.Join(dir, "a/b/c"))
	if len(entries) != 2 {
		t.Errorf("esperava 2 arquivos, tem %d", len(entries))
	}
}

func TestHandlerCSRFMissing(t *testing.T) {
	dir := t.TempDir()
	h := NewHandler(Options{
		StorageRoot: dir,
		MaxUploadMB: 5,
		WebPQuality: 85,
		CSRFCheck:   func(tok string) bool { return tok == "valid" },
	})
	jpg, _ := os.ReadFile("../../testdata/sample.jpg")
	body, ct := makeMultipart(t, "a/b/c", map[string][]byte{"x.jpg": jpg})
	req := httptest.NewRequest(http.MethodPost, "/admin/upload", body)
	req.Header.Set("Content-Type", ct)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}
