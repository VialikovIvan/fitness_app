package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/VialikovIvan/fitness_app/internal/membership/domain"
	"github.com/VialikovIvan/fitness_app/internal/middleware"
	"github.com/shopspring/decimal"
)

const serviceCallTimeout = 3 * time.Second

type MembershipService interface {
	GetPlans(ctx context.Context) ([]domain.Plan, error)
	CreatePlan(ctx context.Context, plan domain.Plan) error
}

type Validator interface {
	Struct(s interface{}) error
}

type Handler struct {
	service  MembershipService
	validate Validator
	logger   *slog.Logger
}

func New(svc MembershipService, v Validator, l *slog.Logger) *Handler {
	return &Handler{
		service:  svc,
		validate: v,
		logger:   l.With("layer", "transport"),
	}
}

type planResponse struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	DurationDays int             `json:"duration_days"`
	Price        decimal.Decimal `json:"price"`
	Currency     string          `json:"currency"`
}

type getPlansResponse struct {
	Plans []planResponse `json:"plans"`
}

type createPlanRequest struct {
	Name         string          `json:"name" validate:"required,max=100"`
	DurationDays int             `json:"duration_days" validate:"gt=0"`
	Price        decimal.Decimal `json:"price" validate:"required"`
	Currency     string          `json:"currency" validate:"required,len=3"`
}

type createPlanResponse struct {
	Message string `json:"message"`
}

func (h *Handler) GetPlans(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}
	if userID, ok := middleware.GetUserID(r); ok {
		logger = h.logger.With("userID", userID)
	}

	ctx, cancel := context.WithTimeout(r.Context(), serviceCallTimeout)
	defer cancel()

	plans, err := h.service.GetPlans(ctx)
	if err != nil {
		handleServiceError(w, err, logger)
		return
	}

	httpPlans := make([]planResponse, 0, len(plans))

	for _, p := range plans {
		httpPlans = append(httpPlans, planResponse{
			ID:           p.ID,
			Name:         p.Name,
			DurationDays: p.DurationDays,
			Price:        p.Price,
			Currency:     p.Currency,
		})
	}

	resp := getPlansResponse{
		Plans: httpPlans,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	logger := h.logger

	if traceID, ok := middleware.GetTraceID(r); ok {
		logger = h.logger.With("traceID", traceID)
	}
	if userID, ok := middleware.GetUserID(r); ok {
		logger = h.logger.With("userID", userID)
	}

	var req createPlanRequest
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

	plan := domain.Plan{
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        req.Price,
		Currency:     req.Currency,
	}

	ctx, cancel := context.WithTimeout(r.Context(), serviceCallTimeout)
	defer cancel()

	if err := h.service.CreatePlan(ctx, plan); err != nil {
		handleServiceError(w, err, logger)
		return
	}

	resp := createPlanResponse{
		Message: "created the plan successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logger.Error("failed to encode response", "err", err)
		return
	}
}

func handleServiceError(w http.ResponseWriter, err error, logger *slog.Logger) {
	switch {
	case errors.Is(err, domain.ErrInvalidPrice):
		logger.Warn("invalid price")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

	case errors.Is(err, domain.ErrInvalidCurrency):
		logger.Warn("invalid currency")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)

	default:
		logger.Error("unexpected error", "err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
