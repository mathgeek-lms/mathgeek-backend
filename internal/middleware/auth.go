package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mathgeek-lms/mathgeek-backend/internal/service"
)

type contextKey string

const claimsKey contextKey = "jwt_claims"

type JWTValidator interface {
	ValidateAccessToken(token string) (*service.Claims, error)
}

func JWTAuth(validator JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeUnauthorized(w, "invalid authorization format: expected 'Bearer <token>'")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])

			claims, err := validator.ValidateAccessToken(tokenStr)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrTokenExpired):
					writeUnauthorized(w, "token expired")
				default:
					writeUnauthorized(w, "invalid token")
				}
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":%q}`, msg)
}

func GetClaims(ctx context.Context) (*service.Claims, bool) {
	c, ok := ctx.Value(claimsKey).(*service.Claims)
	return c, ok
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		roleSet[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok || claims == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if _, allowed := roleSet[claims.Role]; !allowed {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
