package core

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"example.com/pz10-auth/internal/http/middleware"
	"example.com/pz10-auth/internal/platform/config"
	"example.com/pz10-auth/internal/platform/jwt"
	"example.com/pz10-auth/internal/repo"
	"github.com/go-chi/chi/v5"
)

type Service struct {
	repo        *repo.UserMem
	jwt         jwt.Validator
	cfg         config.Config
	blacklist   map[string]time.Time // Отозванные refresh-токены
	mu          sync.Mutex
}

func NewService(r *repo.UserMem, j jwt.Validator, cfg config.Config) *Service {
	return &Service{
		repo: r, jwt: j, cfg: cfg,
		blacklist: make(map[string]time.Time),
	}
}

func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Email == "" || in.Password == "" {
		httpError(w, 400, "invalid_credentials")
		return
	}

	u, err := s.repo.CheckPassword(in.Email, in.Password)
	if err != nil {
		httpError(w, 401, "unauthorized")
		return
	}

	access, _ := s.jwt.Sign(u.ID, u.Email, u.Role, "access", s.cfg.JWTTTL)
	refresh, _ := s.jwt.Sign(u.ID, u.Email, u.Role, "refresh", s.cfg.RefreshTTL)

	jsonOK(w, map[string]any{
		"access_token": access,
		"refresh_token": refresh,
	})
}

func (s *Service) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	var in struct{ RefreshToken string `json:"refresh_token"` }
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.RefreshToken == "" {
		httpError(w, 400, "invalid_request")
		return
	}

	// Проверка блэклиста
	s.mu.Lock()
	_, isRevoked := s.blacklist[in.RefreshToken]
	s.mu.Unlock()
	if isRevoked {
		httpError(w, 401, "token_revoked")
		return
	}

	// Валидация токена
	claims, err := s.jwt.Parse(in.RefreshToken, "refresh")
	if err != nil {
		httpError(w, 401, "invalid_refresh_token")
		return
	}

	// Извлекаем данные из claims
	sub := int64(claims["sub"].(float64))
	email := claims["email"].(string)
	role := claims["role"].(string)

	// Генерируем новые токены
	newAccess, _ := s.jwt.Sign(sub, email, role, "access", s.cfg.JWTTTL)
	newRefresh, _ := s.jwt.Sign(sub, email, role, "refresh", s.cfg.RefreshTTL)

	// Добавляем старый в блэклист (отзываем)
	exp := int64(claims["exp"].(float64))
	s.mu.Lock()
	s.blacklist[in.RefreshToken] = time.Unix(exp, 0)
	s.mu.Unlock()

	jsonOK(w, map[string]any{
		"access_token": newAccess,
		"refresh_token": newRefresh,
	})
}

func (s *Service) MeHandler(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r) // ← было GetClaimsCtx(r)
	jsonOK(w, map[string]any{
		"id":    claims["sub"],
		"email": claims["email"],
		"role":  claims["role"],
	})
}

func (s *Service) AdminStats(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]any{"users": 2, "version": "1.0"})
}

func (s *Service) UserProfile(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	jsonOK(w, map[string]any{
		"message":      "Profile data accessed successfully",
		"requested_id": chi.URLParam(r, "id"), // ← не r.PathValue!
		"your_sub":     claims["sub"],
	})
}

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func httpError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}