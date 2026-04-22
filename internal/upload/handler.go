package upload

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	StorageRoot   string
	MaxUploadMB   int
	WebPQuality   int
	PublicBaseURL string
	CSRFCheck     func(string) bool
	Logger        *slog.Logger
}

type Handler struct {
	opts Options
}

func NewHandler(o Options) *Handler {
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	if o.CSRFCheck == nil {
		o.CSRFCheck = func(string) bool { return false }
	}
	return &Handler{opts: o}
}

func (h *Handler) WithCSRFCheck(check func(string) bool) *Handler {
	opts := h.opts
	opts.CSRFCheck = check
	return &Handler{opts: opts}
}

type uploadResult struct {
	OriginalFilename string `json:"original_filename"`
	FinalFilename    string `json:"final_filename"`
	URL              string `json:"url"`
	ImgTag           string `json:"img_tag"`
	Renamed          bool   `json:"renamed"`
}

type uploadResponse struct {
	Results []uploadResult `json:"results"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	maxBytes := int64(h.opts.MaxUploadMB) * 1024 * 1024
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes*20)

	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "form inválida: "+err.Error(), http.StatusBadRequest)
		return
	}

	if !h.opts.CSRFCheck(r.FormValue("csrf")) {
		http.Error(w, "csrf inválido", http.StatusForbidden)
		return
	}

	rawPath := r.FormValue("path")
	sanitizedPath := SanitizePath(rawPath)
	if !IsValidPath(sanitizedPath) {
		http.Error(w, "path inválido", http.StatusBadRequest)
		return
	}

	targetDir := filepath.Join(h.opts.StorageRoot, sanitizedPath)
	storageAbs, err := filepath.Abs(h.opts.StorageRoot)
	if err != nil {
		http.Error(w, "storage root", http.StatusInternalServerError)
		return
	}
	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		http.Error(w, "target dir", http.StatusInternalServerError)
		return
	}
	if !strings.HasPrefix(targetAbs+string(os.PathSeparator), storageAbs+string(os.PathSeparator)) && targetAbs != storageAbs {
		http.Error(w, "path fora do storage", http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(targetAbs, 0750); err != nil {
		http.Error(w, "mkdir: "+err.Error(), http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "nenhum arquivo", http.StatusBadRequest)
		return
	}

	results := make([]uploadResult, 0, len(files))
	for _, fh := range files {
		if fh.Size > maxBytes {
			http.Error(w, fmt.Sprintf("arquivo %s excede %d MB", fh.Filename, h.opts.MaxUploadMB), http.StatusRequestEntityTooLarge)
			return
		}

		f, err := fh.Open()
		if err != nil {
			http.Error(w, "open: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			http.Error(w, "read: "+err.Error(), http.StatusInternalServerError)
			return
		}

		sniffedMIME := http.DetectContentType(data)
		if !isAcceptedMIME(sniffedMIME) {
			h.opts.Logger.Warn("mime rejected",
				slog.String("filename", fh.Filename),
				slog.String("sniffed", sniffedMIME),
				slog.String("remote", r.RemoteAddr),
			)
			http.Error(w, "mime não suportado: "+sniffedMIME, http.StatusUnsupportedMediaType)
			return
		}

		webpBytes, err := ConvertToWebP(data, sniffedMIME, h.opts.WebPQuality)
		if err != nil {
			http.Error(w, "conversão webp: "+err.Error(), http.StatusInternalServerError)
			return
		}

		finalName := replaceExtToWebP(SanitizeFilename(fh.Filename))
		fullPath := filepath.Join(targetAbs, finalName)
		renamed := false
		if _, err := os.Stat(fullPath); err == nil {
			stamp := time.Now().UTC().Format("20060102-1504")
			base := strings.TrimSuffix(finalName, ".webp")
			finalName = fmt.Sprintf("%s_%s.webp", base, stamp)
			fullPath = filepath.Join(targetAbs, finalName)
			renamed = true
		}

		if err := writeAtomic(fullPath, webpBytes); err != nil {
			http.Error(w, "write: "+err.Error(), http.StatusInternalServerError)
			return
		}

		urlPath := sanitizedPath + "/" + finalName
		publicURL := strings.TrimRight(h.opts.PublicBaseURL, "/") + "/" + urlPath
		results = append(results, uploadResult{
			OriginalFilename: fh.Filename,
			FinalFilename:    finalName,
			URL:              publicURL,
			ImgTag:           fmt.Sprintf(`<img src="%s" />`, publicURL),
			Renamed:          renamed,
		})

		h.opts.Logger.Info("upload ok",
			slog.String("remote", r.RemoteAddr),
			slog.String("path", urlPath),
			slog.Int("size_original", len(data)),
			slog.Int("size_final", len(webpBytes)),
			slog.String("mime", sniffedMIME),
			slog.Bool("renamed", renamed),
		)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(uploadResponse{Results: results})
}

func isAcceptedMIME(mime string) bool {
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
		return true
	}
	return false
}

func replaceExtToWebP(name string) string {
	ext := filepath.Ext(name)
	if ext == "" {
		return name + ".webp"
	}
	return strings.TrimSuffix(name, ext) + ".webp"
}

func writeAtomic(dst string, data []byte) error {
	tmp := dst + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, dst)
}
