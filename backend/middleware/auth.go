package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pointofsale/backend/utils"
)

type userContextKey string

const UserClaimsKey userContextKey = "user_claims"

func Auth(accessSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				utils.Error(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := utils.ValidateToken(parts[1], accessSecret)
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) *utils.Claims {
	if claims, ok := ctx.Value(UserClaimsKey).(*utils.Claims); ok {
		return claims
	}
	return nil
}
