package auth

import (
	"context"
	"crypto/subtle"

	"github.com/arxdsilva/opencoverage/internal/domain"
)

type EnvAPIKeyAuthenticator struct {
	expectedSecret string
}

func NewEnvAPIKeyAuthenticator(expectedSecret string) *EnvAPIKeyAuthenticator {
	return &EnvAPIKeyAuthenticator{expectedSecret: expectedSecret}
}

func (a *EnvAPIKeyAuthenticator) Authenticate(_ context.Context, apiKey string) error {
	if subtle.ConstantTimeCompare([]byte(apiKey), []byte(a.expectedSecret)) != 1 {
		return domain.ErrNotFound
	}

	return nil
}

func (a *EnvAPIKeyAuthenticator) WantedAPIKey() string {
	return a.expectedSecret
}
