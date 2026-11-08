package middleware

import (
	"context"
	"net/http"
	"strings"

	"example.com/pz10-auth/internal/platform/jwt"
)

type ctxKey int
const ctxClaimsKey ctxKey = iota

// Экспортируемый геттер — использует тот же ключ, что и AuthN
func GetClaims(r *http.Request) map[string]any {
	claims, _ := r.Context().Value(ctxClaimsKey).(map[string]any)
	return claims
}

func AuthN(v jwt.Validator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			raw := strings.TrimPrefix(h, "Bearer ")
			claims, err := v.Parse(raw, "access")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ctxClaimsKey, map[string]any(claims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}