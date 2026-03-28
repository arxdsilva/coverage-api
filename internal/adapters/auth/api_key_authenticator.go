package auth

import (
	"context"
	"crypto/subtle"

	"github.com/arxdsilva/coverage-api/internal/domain"
)

type EnvAPIKeyAuthenticator struct {
	expectedSecret string
}

func NewEnvAPIKeyAuthenticator(expectedSecret string) *EnvAPIKeyAuthenticator {
	return &EnvAPIKeyAuthenticator{expectedSecret: expectedSecret}
}

func (a *EnvAPIKeyAuthenticator) Authenticate(ctx context.Context, apiKey string) error {
	_ = ctx
	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(a.expectedSecret)) != 1 {
		return domain.ErrNotFound
	}

	return nil
}
