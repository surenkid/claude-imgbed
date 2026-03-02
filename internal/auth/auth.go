package auth

import (
	"strings"
)

type Authenticator struct {
	token string
}

func NewAuthenticator(token string) *Authenticator {
	return &Authenticator{
		token: token,
	}
}

func (a *Authenticator) ValidateToken(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	return parts[1] == a.token
}
