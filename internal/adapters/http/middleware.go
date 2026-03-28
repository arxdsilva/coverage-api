package http

import (
	"net/http"
	"strings"

	"github.com/arxdsilva/coverage-api/internal/application"
	"github.com/arxdsilva/coverage-api/internal/domain"
)

func APIKeyMiddleware(auth application.APIKeyAuthenticator, headerName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := strings.TrimSpace(r.Header.Get(headerName))
			if key == "" {
				writeError(w, http.StatusUnauthorized, &application.AppError{
					Code:    application.CodeUnauthenticated,
					Message: "missing API key",
					Details: map[string]any{"header": headerName},
				})
				return
			}

			if err := auth.Authenticate(r.Context(), key); err != nil {
				if err == domain.ErrNotFound {
					writeError(w, http.StatusUnauthorized, &application.AppError{
						Code:    application.CodeUnauthenticated,
						Message: "invalid API key",
					})
					return
				}
				writeError(w, http.StatusInternalServerError, &application.AppError{
					Code:    application.CodeInternal,
					Message: "failed to authenticate API key",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
