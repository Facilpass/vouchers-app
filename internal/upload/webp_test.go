package upload

import (
	"bytes"
	"os"
	"testing"
)

func TestConvertJPEGToWebP(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample.jpg")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	out, err := ConvertToWebP(data, "image/jpeg", 85)
	if err != nil {
		t.Fatalf("ConvertToWebP: %v", err)
	}
	if !bytes.Equal(out[:4], []byte("RIFF")) {
		t.Errorf("output não inicia com RIFF: %x", out[:4])
	}
	if !bytes.Equal(out[8:12], []byte("WEBP")) {
		t.Errorf("output não contém WEBP magic: %x", out[8:12])
	}
}

func TestConvertPNGToWebP(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample.png")
	if err != nil {
		t.Fatalf("fixture: %v", err)
	}
	out, err := ConvertToWebP(data, "image/png", 85)
	if err != nil {
		t.Fatalf("ConvertToWebP: %v", err)
	}
	if !bytes.Equal(out[:4], []byte("RIFF")) {
		t.Errorf("output não inicia com RIFF")
	}
}

func TestConvertWebPPassthrough(t *testing.T) {
	data := []byte("RIFF\x00\x00\x00\x00WEBPVP8 ")
	out, err := ConvertToWebP(data, "image/webp", 85)
	if err != nil {
		t.Fatalf("WebP passthrough: %v", err)
	}
	if !bytes.Equal(out, data) {
		t.Error("WebP input deveria passar direto (bytes iguais)")
	}
}

func TestConvertInvalidMIME(t *testing.T) {
	_, err := ConvertToWebP([]byte("fake"), "application/pdf", 85)
	if err == nil {
		t.Fatal("MIME não suportado deveria falhar")
	}
}
