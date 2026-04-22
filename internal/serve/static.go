package serve

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Handler struct {
	root string
}

func NewHandler(root string) *Handler {
	abs, err := filepath.Abs(root)
	if err != nil {
		abs = filepath.Clean(root)
	}
	return &Handler{root: abs}
}

var allowedExts = map[string]string{
	".webp": "image/webp",
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if strings.HasSuffix(r.URL.Path, "/") {
		http.NotFound(w, r)
		return
	}

	urlPath := strings.TrimPrefix(r.URL.Path, "/")
	if urlPath == "" {
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(filepath.Ext(urlPath))
	ct, ok := allowedExts[ext]
	if !ok {
		http.NotFound(w, r)
		return
	}

	full := filepath.Join(h.root, urlPath)
	resolved, err := filepath.Abs(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if !strings.HasPrefix(resolved+string(os.PathSeparator), h.root+string(os.PathSeparator)) && resolved != h.root {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(resolved)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", "inline")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	http.ServeFile(w, r, resolved)
}
