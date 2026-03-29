package main

import (
	"embed"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed web/*
var embeddedFrontend embed.FS

type config struct {
	Addr         string
	APIBaseURL   string
	APIKeyHeader string
	APIKeySecret string
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := loadConfig()

	frontendFS, err := fs.Sub(embeddedFrontend, "web")
	if err != nil {
		slog.Error("frontend_start_failed", "stage", "load_static_files", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/assets/", http.StripPrefix("/", http.FileServer(http.FS(frontendFS))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveEmbeddedFile(w, http.FS(frontendFS), "index.html")
	})

	mux.HandleFunc("/api/projects", proxyHandler(cfg))
	mux.HandleFunc("/api/projects/", proxyHandler(cfg))

	server := &http.Server{
		Addr:         cfg.Addr,
		Handler:      requestLogger(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("frontend_starting", "addr", cfg.Addr, "api_base_url", cfg.APIBaseURL)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("frontend_server_failed", "error", err)
		os.Exit(1)
	}
}

func loadConfig() config {
	return config{
		Addr:         envOrDefault("FRONTEND_ADDR", ":8090"),
		APIBaseURL:   strings.TrimRight(envOrDefault("API_BASE_URL", "http://localhost:8080"), "/"),
		APIKeyHeader: envOrDefault("API_KEY_HEADER", "X-API-Key"),
		APIKeySecret: envOrDefault("API_KEY_SECRET", "dev-local-key"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func proxyHandler(cfg config) http.HandlerFunc {
	client := &http.Client{Timeout: 20 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		target := cfg.APIBaseURL + "/v1" + strings.TrimPrefix(r.URL.Path, "/api")
		u, err := url.Parse(target)
		if err != nil {
			http.Error(w, "invalid target url", http.StatusInternalServerError)
			return
		}
		u.RawQuery = r.URL.RawQuery

		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u.String(), nil)
		if err != nil {
			http.Error(w, "failed to build request", http.StatusInternalServerError)
			return
		}
		req.Header.Set(cfg.APIKeyHeader, cfg.APIKeySecret)

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "upstream request failed", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		if ct := resp.Header.Get("Content-Type"); ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("frontend_request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(start).Milliseconds())
	})
}

func serveEmbeddedFile(w http.ResponseWriter, fsys http.FileSystem, name string) {
	f, err := fsys.Open(name)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	_, _ = io.Copy(w, f)
}
