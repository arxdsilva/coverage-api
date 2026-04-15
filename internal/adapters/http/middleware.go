package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/arxdsilva/opencoverage/internal/application"
	"github.com/arxdsilva/opencoverage/internal/domain"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rw *statusRecorder) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestLoggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			slog.Info("http_request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", chiMiddleware.GetReqID(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

func APIKeyMiddleware(auth application.APIKeyAuthenticator, headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := chiMiddleware.GetReqID(r.Context())
			key := strings.TrimSpace(r.Header.Get(headerName))
			if key == "" {
				slog.Warn("auth_check",
					"method", r.Method,
					"path", r.URL.Path,
					"result", "missing_api_key",
					"request_id", requestID,
				)
				writeError(w, http.StatusUnauthorized, &application.AppError{
					Code:    application.CodeUnauthenticated,
					Message: "missing API key",
					Details: map[string]any{"header": headerName},
				})
				return
			}

			if err := auth.Authenticate(r.Context(), key); err != nil {
				if err == domain.ErrNotFound {
					slog.Warn("auth_check",
						"method", r.Method,
						"path", r.URL.Path,
						"result", "invalid_api_key",
						"request_id", requestID,
						"wanted_api_key", auth.WantedAPIKey(),
					)
					writeError(w, http.StatusUnauthorized, &application.AppError{
						Code:    application.CodeUnauthenticated,
						Message: "invalid API key",
					})
					return
				}
				slog.Error("auth_check",
					"method", r.Method,
					"path", r.URL.Path,
					"result", "auth_error",
					"request_id", requestID,
					"error", err,
				)
				writeError(w, http.StatusInternalServerError, &application.AppError{
					Code:    application.CodeInternal,
					Message: "failed to authenticate API key",
				})
				return
			}

			slog.Info("auth_check",
				"method", r.Method,
				"path", r.URL.Path,
				"result", "ok",
				"request_id", requestID,
			)

			next.ServeHTTP(w, r)
		})
	}
}
