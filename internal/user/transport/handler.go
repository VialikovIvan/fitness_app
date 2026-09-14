package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/middleware"
	"github.com/VialikovIvan/fitness_app/internal/user/domain"
)

const cookieValidityTime = 24 * time.Hour

type AuthService interface {
	Register(ctx context.Context, user domain.User, plainPassword string) (string, error)
	Login(ctx context.Context, user domain.User, plainPassword string) (string, error)
}

type Validator interface {
	Struct(s interface{}) error
}

type Handler struct {
	service  AuthService
	validate Validator
	logger   *slog.Logger
}

func New(svc AuthService, v Validator, l *slog.Logger) *Handler {
	return &Handler{
		service:  svc,
		validate: v,
		logger:   l.With("layer", "transport"),
	}
}

type registerRequest struct {
	FirstName  string `json:"first_name" validate:"required,min=2,max=100"`
	MiddleName string `json:"middle_name" validate:"max=100"`
	LastName   string `json:"last_name" validate:"required,min=2,max=100"`
	Password   string `json:"password" validate:"required,min=8,max=72"`
	Email      string `json:"email" validate:"required,email,max=255"`
}

type loginRequest struct {
	Password string `json:"password" validate:"required,min=8,max=72"`
	Email    string `json:"email" validate:"required,email,max=255"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

type logoutResponse struct {
	Message string `json:"message"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}

	logger.Info("register request received")

	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn("failed to decode request", "err", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(req)
	if err != nil {
		logger.Warn("received invalid data", "err", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	user := domain.User{
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		Email:      req.Email,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	token, err := h.service.Register(ctx, user, req.Password)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}

	cookie := http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(cookieValidityTime),
	}

	http.SetCookie(w, &cookie)

	resp := tokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}

	logger.Info("user registered successfully")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}

	logger.Info("login request received")

	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Warn("failed to decode request", "err", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(req)
	if err != nil {
		logger.Warn("received invalid data", "err", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	user := domain.User{
		Email: req.Email,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	token, err := h.service.Login(ctx, user, req.Password)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}

	cookie := http.Cookie{
		Name:     "authToken",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(cookieValidityTime),
	}

	http.SetCookie(w, &cookie)

	resp := tokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}

	logger.Info("user logged in successfully")
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}

	logger.Info("logout request received")

	cookie := http.Cookie{
		Name:     "authToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	}

	http.SetCookie(w, &cookie)

	resp := logoutResponse{
		Message: "logged out successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}

	logger.Info("user logged out successfully")
}

func handleServiceError(w http.ResponseWriter, err error, logger *slog.Logger) {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		logger.Warn("user already exists")
		http.Error(w, err.Error(), http.StatusConflict)

	case errors.Is(err, domain.ErrUserDoesNotExist):
		logger.Warn("user does not exist")
		http.Error(w, err.Error(), http.StatusUnauthorized)

	case errors.Is(err, domain.ErrInvalidPassword):
		logger.Warn("invalid password")
		http.Error(w, err.Error(), http.StatusUnauthorized)

	case errors.Is(err, domain.ErrInvalidPasswordLength):
		logger.Warn("invalid password length")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

	case errors.Is(err, domain.ErrInvalidEmailFormat):
		logger.Warn("invalid email format")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

	default:
		logger.Error("unexpected error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Only commit in main branch
