package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"example.com/pz9-auth/internal/core"
	"example.com/pz9-auth/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

const maxRequestBody = 1 << 20 // 1 MiB
const maxBcryptPasswordBytes = 72

type UserStore interface {
	Create(ctx context.Context, user *core.User) error
	ByEmail(ctx context.Context, email string) (core.User, error)
}

type AuthHandler struct {
	Users      UserStore
	BcryptCost int
}

type credentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type responseUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type authResponse struct {
	Status string        `json:"status"`
	User   *responseUser `json:"user,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	input.Email = normalizeEmail(input.Email)
	if !validEmail(input.Email) || utf8.RuneCountInString(input.Password) < 8 || len(input.Password) > maxBcryptPasswordBytes {
		writeErr(w, http.StatusBadRequest, "invalid_email_or_password")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), h.BcryptCost)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "hash_failed")
		return
	}

	user := core.User{
		Email:        input.Email,
		PasswordHash: string(hash),
	}
	if err := h.Users.Create(r.Context(), &user); err != nil {
		if errors.Is(err, repo.ErrEmailTaken) {
			writeErr(w, http.StatusConflict, "email_taken")
			return
		}
		writeErr(w, http.StatusInternalServerError, "db_error")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Status: "ok",
		User:   &responseUser{ID: user.ID, Email: user.Email},
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if err := decodeJSON(w, r, &input); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json")
		return
	}

	input.Email = normalizeEmail(input.Email)
	if !validEmail(input.Email) || input.Password == "" || len(input.Password) > maxBcryptPasswordBytes {
		writeErr(w, http.StatusBadRequest, "email_and_password_required")
		return
	}

	user, err := h.Users.ByEmail(r.Context(), input.Email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			writeErr(w, http.StatusUnauthorized, "invalid_credentials")
			return
		}
		writeErr(w, http.StatusInternalServerError, "db_error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		writeErr(w, http.StatusUnauthorized, "invalid_credentials")
		return
	}

	// JWT issuance is intentionally postponed to practical assignment No. 10.
	writeJSON(w, http.StatusOK, authResponse{
		Status: "ok",
		User:   &responseUser{ID: user.ID, Email: user.Email},
	})
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	if email == "" || len(email) > 255 {
		return false
	}
	parsed, err := mail.ParseAddress(email)
	return err == nil && parsed.Address == email
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain exactly one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeErr(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
