package main

import (
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	vouchersapp "gitea.homelab.local/facilpass/vouchers-app"
	"gitea.homelab.local/facilpass/vouchers-app/internal/auth"
	"gitea.homelab.local/facilpass/vouchers-app/internal/config"
	"gitea.homelab.local/facilpass/vouchers-app/internal/serve"
	"gitea.homelab.local/facilpass/vouchers-app/internal/ui"
	"gitea.homelab.local/facilpass/vouchers-app/internal/upload"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("config inválida", slog.String("err", err.Error()))
		os.Exit(1)
	}
	logger.Info("config loaded",
		slog.String("listen", cfg.ListenAddr),
		slog.String("storage", cfg.StorageRoot),
		slog.Int("max_mb", cfg.MaxUploadMB),
		slog.Int("webp_quality", cfg.WebPQuality),
		slog.String("public_url", cfg.PublicBaseURL),
	)

	if err := os.MkdirAll(cfg.StorageRoot, 0750); err != nil {
		logger.Error("mkdir storage", slog.String("err", err.Error()))
		os.Exit(1)
	}

	tplFS, err := fs.Sub(vouchersapp.TemplatesFS, "web")
	if err != nil {
		logger.Error("template fs sub", slog.String("err", err.Error()))
		os.Exit(1)
	}
	staticSub, err := fs.Sub(vouchersapp.StaticFS, "web/static")
	if err != nil {
		logger.Error("static fs sub", slog.String("err", err.Error()))
		os.Exit(1)
	}

	renderer, err := ui.NewTemplateRenderer(tplFS)
	if err != nil {
		logger.Error("render init", slog.String("err", err.Error()))
		os.Exit(1)
	}

	sessions := auth.NewSessionManager(cfg.SessionSecret, 8*time.Hour)
	rateLimiter := auth.NewLoginRateLimiter(5, 15*time.Minute)
	csrf := ui.NewCSRFManager(cfg.SessionSecret)

	uiHandler := ui.New(ui.UIOptions{
		AdminUser:       cfg.AdminUser,
		AdminPassBcrypt: cfg.AdminPassBcrypt,
		Sessions:        sessions,
		RateLimiter:     rateLimiter,
		CSRF:            csrf,
		Templates:       renderer,
		Logger:          logger,
		CookieSecure:    strings.HasPrefix(cfg.PublicBaseURL, "https://"),
	})

	uploadHandler := upload.NewHandler(upload.Options{
		StorageRoot:   cfg.StorageRoot,
		MaxUploadMB:   cfg.MaxUploadMB,
		WebPQuality:   cfg.WebPQuality,
		PublicBaseURL: cfg.PublicBaseURL,
		Logger:        logger,
	})

	serveHandler := serve.NewHandler(cfg.StorageRoot)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintln(w, "ok")
	})

	mux.Handle("GET /admin/static/", http.StripPrefix("/admin/static/", http.FileServer(http.FS(staticSub))))

	mux.HandleFunc("GET /admin", uiHandler.GetLogin)
	mux.HandleFunc("GET /admin/", uiHandler.GetLogin)
	mux.HandleFunc("POST /admin/login", uiHandler.PostLogin)
	mux.HandleFunc("GET /admin/logout", uiHandler.GetLogout)

	mux.Handle("GET /admin/app", uiHandler.AuthMiddleware(http.HandlerFunc(uiHandler.GetApp)))
	mux.Handle("POST /admin/upload", uiHandler.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadHandler.WithCSRFCheck(uiHandler.CSRFCheckFromRequest(r)).ServeHTTP(w, r)
	})))

	mux.Handle("GET /", serveHandler)

	logger.Info("starting server", slog.String("addr", cfg.ListenAddr))
	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       180 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server exit", slog.String("err", err.Error()))
		os.Exit(1)
	}
}
