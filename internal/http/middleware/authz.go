package middleware

import (
	"net/http"
	"fmt"
	"github.com/go-chi/chi/v5"
)

func AuthZRoles(allowed ...string) func(http.Handler) http.Handler {
	set := map[string]struct{}{}
	for _, a := range allowed { set[a] = struct{}{} }
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			role, _ := claims["role"].(string)
			if _, ok := set[role]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ABAC: user может смотреть только свой профиль. Admin может смотреть всех.
func AuthZABAC() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			role, _ := claims["role"].(string)
			
			// Если админ - пропускаем
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Если юзер - проверяем id
			urlID := chi.URLParam(r, "id")
			subStr := fmt.Sprintf("%v", claims["sub"])
			
			if urlID != subStr {
				http.Error(w, "forbidden: access to another user profile", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}