package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"example.com/pz9-auth/internal/core"
	"example.com/pz9-auth/internal/repo"
	"golang.org/x/crypto/bcrypt"
)

type memoryUserStore struct {
	mu        sync.Mutex
	users     map[string]core.User
	nextID    int64
	createErr error
	lookupErr error
}

func newMemoryUserStore() *memoryUserStore {
	return &memoryUserStore{users: make(map[string]core.User), nextID: 1}
}

func (s *memoryUserStore) Create(_ context.Context, user *core.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.createErr != nil {
		return s.createErr
	}
	if _, exists := s.users[user.Email]; exists {
		return repo.ErrEmailTaken
	}
	user.ID = s.nextID
	s.nextID++
	s.users[user.Email] = *user
	return nil
}

func (s *memoryUserStore) ByEmail(_ context.Context, email string) (core.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.lookupErr != nil {
		return core.User{}, s.lookupErr
	}
	user, exists := s.users[email]
	if !exists {
		return core.User{}, repo.ErrUserNotFound
	}
	return user, nil
}

func newTestHandler(store UserStore) *AuthHandler {
	return &AuthHandler{Users: store, BcryptCost: bcrypt.MinCost}
}

func performJSONRequest(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}

func TestRegisterCreatesPasswordHash(t *testing.T) {
	store := newMemoryUserStore()
	handler := newTestHandler(store)
	response := performJSONRequest(handler.Register, `{"email":" User@Example.com ","password":"Secret123!"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, response.Code, response.Body.String())
	}

	stored, err := store.ByEmail(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("user was not stored: %v", err)
	}
	if stored.PasswordHash == "Secret123!" {
		t.Fatal("password was stored as plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("Secret123!")); err != nil {
		t.Fatalf("stored hash does not match original password: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, exists := body["passwordHash"]; exists {
		t.Fatal("API response must not expose passwordHash")
	}
}

func TestRegisterDuplicateEmailReturnsConflict(t *testing.T) {
	store := newMemoryUserStore()
	handler := newTestHandler(store)
	body := `{"email":"user@example.com","password":"Secret123!"}`
	first := performJSONRequest(handler.Register, body)
	second := performJSONRequest(handler.Register, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("first registration: expected 201, got %d", first.Code)
	}
	if second.Code != http.StatusConflict {
		t.Fatalf("second registration: expected 409, got %d", second.Code)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	handler := newTestHandler(newMemoryUserStore())
	for _, body := range []string{
		`{"email":"not-an-email","password":"Secret123!"}`,
		`{"email":"user@example.com","password":"short"}`,
		`{"email":"user@example.com","password":"Secret123!","extra":true}`,
		`{broken json`,
	} {
		response := performJSONRequest(handler.Register, body)
		if response.Code != http.StatusBadRequest {
			t.Errorf("body %q: expected 400, got %d", body, response.Code)
		}
	}
}

func TestLoginSuccessAndInvalidCredentials(t *testing.T) {
	store := newMemoryUserStore()
	handler := newTestHandler(store)
	registration := performJSONRequest(handler.Register, `{"email":"user@example.com","password":"Secret123!"}`)
	if registration.Code != http.StatusCreated {
		t.Fatalf("setup registration failed: %d", registration.Code)
	}

	success := performJSONRequest(handler.Login, `{"email":"user@example.com","password":"Secret123!"}`)
	if success.Code != http.StatusOK {
		t.Fatalf("expected successful login status 200, got %d", success.Code)
	}

	wrongPassword := performJSONRequest(handler.Login, `{"email":"user@example.com","password":"WrongPass123!"}`)
	unknownEmail := performJSONRequest(handler.Login, `{"email":"unknown@example.com","password":"WrongPass123!"}`)
	if wrongPassword.Code != http.StatusUnauthorized || unknownEmail.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for both invalid credentials, got %d and %d", wrongPassword.Code, unknownEmail.Code)
	}
	if wrongPassword.Body.String() != unknownEmail.Body.String() {
		t.Fatalf("error response leaks whether email exists: %q vs %q", wrongPassword.Body.String(), unknownEmail.Body.String())
	}
}

func TestLoginDatabaseFailureReturnsInternalServerError(t *testing.T) {
	store := newMemoryUserStore()
	store.lookupErr = errors.New("database unavailable")
	handler := newTestHandler(store)
	response := performJSONRequest(handler.Login, `{"email":"user@example.com","password":"Secret123!"}`)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
}
