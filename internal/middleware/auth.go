package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/mathgeek-lms/mathgeek-backend/internal/common"
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
				common.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				common.WriteError(w, http.StatusUnauthorized, "invalid authorization format: expected 'Bearer <token>'")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])

			claims, err := validator.ValidateAccessToken(tokenStr)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrTokenExpired):
					common.WriteError(w, http.StatusUnauthorized, "token expired")
				default:
					common.WriteError(w, http.StatusUnauthorized, "invalid token")
				}
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
				common.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if _, allowed := roleSet[claims.Role]; !allowed {
				common.WriteError(w, http.StatusForbidden, "forbidden")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
