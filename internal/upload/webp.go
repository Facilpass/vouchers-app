package upload

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"

	"github.com/chai2010/webp"
)

// ConvertToWebP decodifica JPEG/PNG e re-codifica como WebP lossy.
// WebP input passa direto sem re-encode.
func ConvertToWebP(data []byte, sniffedMIME string, quality int) ([]byte, error) {
	if quality < 1 || quality > 100 {
		return nil, fmt.Errorf("quality inválida: %d", quality)
	}

	switch sniffedMIME {
	case "image/webp":
		return data, nil
	case "image/jpeg":
		img, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decode jpeg: %w", err)
		}
		return encodeWebP(img, quality)
	case "image/png":
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("decode png: %w", err)
		}
		return encodeWebP(img, quality)
	default:
		return nil, fmt.Errorf("MIME não suportado: %s", sniffedMIME)
	}
}

func encodeWebP(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	opts := &webp.Options{
		Lossless: false,
		Quality:  float32(quality),
	}
	if err := webp.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
}
