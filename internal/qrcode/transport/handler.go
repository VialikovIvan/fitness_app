package transport

import (
	"context"
	"errors"
	"image/png"
	"log/slog"
	"net/http"
	"time"

	mbspDomain "github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	qrDomain "github.com/VialikovIvan/fitness_app/internal/qrcode/domain"
	"github.com/skip2/go-qrcode"
)

const (
	defaultServiceCallTimeout     = 3 * time.Second
	qrCodeSize                int = 256
)

type QrCodeService interface {
	Generate(ctx context.Context, userID int64) (*qrcode.QRCode, error)
	AllowAccess(ctx context.Context, token string) (bool, error)
}

type Handler struct {
	service QrCodeService
	logger  *slog.Logger
}

func New(svc QrCodeService, l *slog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  l.With("layer", "transport"),
	}
}

func (h *Handler) GetQR(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}

	userID, ok := middleware.GetUserID(r)
	if !ok {
		logger.Error("expected user id to exist")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger = h.logger.With("userID", userID)

	ctx, cancel := context.WithTimeout(r.Context(), defaultServiceCallTimeout)
	defer cancel()

	qrCode, err := h.service.Generate(ctx, userID)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	if err := png.Encode(w, qrCode.Image(qrCodeSize)); err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}
}

func (h *Handler) VerifyAccess(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		logger.Error("expected token to exist")
		http.Error(w, "No Token Found", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultServiceCallTimeout)
	defer cancel()

	allowed, err := h.service.AllowAccess(ctx, token)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}
	if !allowed {
		logger.Warn("access denied")
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleServiceError(w http.ResponseWriter, err error, logger *slog.Logger) {
	switch {
	case errors.Is(err, mbspDomain.ErrUserDoesNotHaveActiveMembership):
		logger.Warn("user does not have an active membership")
		http.Error(w, err.Error(), http.StatusForbidden)

	case errors.Is(err, qrDomain.ErrUserIsInCooldown):
		logger.Warn("user is in cooldown")
		http.Error(w, err.Error(), http.StatusForbidden)

	default:
		logger.Error("unexpected error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
