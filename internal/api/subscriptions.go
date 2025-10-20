package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscriptions-service/internal/subscription"
)

// SubscriptionHandler handles HTTP requests for subscriptions.
type SubscriptionHandler struct {
	repo *subscription.Repository
}

// NewSubscriptionHandler creates a new handler with the provided repository.
func NewSubscriptionHandler(repo *subscription.Repository) *SubscriptionHandler {
	return &SubscriptionHandler{repo: repo}
}

// RegisterRoutes registers subscription routes on the provided router.
func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Get("/", h.List)
		r.Get("", h.List)
		r.Post("/", h.Create)
		r.Post("", h.Create)
		r.Get("/summary", h.Summary)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Get("", h.Get)
			r.Put("/", h.Update)
			r.Put("", h.Update)
			r.Delete("/", h.Delete)
			r.Delete("", h.Delete)
		})
	})
}

type createSubscriptionRequest struct {
	UserID        string `json:"user_id"`
	Plan          string `json:"plan"`
	AmountCents   int64  `json:"amount_cents"`
	Currency      string `json:"currency"`
	BillingPeriod string `json:"billing_period"`
}

func decodeRequest(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.mapRequestToModel(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.repo.Create(r.Context(), sub); err != nil {
		http.Error(w, "failed to create subscription", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, h.mapModelToResponse(sub))
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	subs, err := h.repo.List(r.Context())
	if err != nil {
		http.Error(w, "failed to list subscriptions", http.StatusInternalServerError)
		return
	}

	responses := make([]subscriptionResponse, 0, len(subs))
	for _, s := range subs {
		responses = append(responses, h.mapModelToResponse(&s))
	}
	writeJSON(w, http.StatusOK, responses)
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	sub, err := h.repo.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to fetch subscription", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, h.mapModelToResponse(sub))
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	var req createSubscriptionRequest
	if err := decodeRequest(r, &req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub, err := h.mapRequestToModel(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sub.ID = id

	if err := h.repo.Update(r.Context(), sub); err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to update subscription", http.StatusInternalServerError)
		return
	}

	// Fetch latest view to include updated_at set by trigger.
	updated, err := h.repo.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to fetch subscription", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, h.mapModelToResponse(updated))
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to delete subscription", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type summaryResponse struct {
	Month            string `json:"month"`
	TotalAmountCents int64  `json:"total_amount_cents"`
}

func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	monthParam := r.URL.Query().Get("month")
	if monthParam == "" {
		http.Error(w, "month query parameter is required", http.StatusBadRequest)
		return
	}

	period, err := parseMonth(monthParam)
	if err != nil {
		http.Error(w, "invalid month format", http.StatusBadRequest)
		return
	}

	total, err := h.repo.MonthlySummary(r.Context(), period)
	if err != nil {
		http.Error(w, "failed to fetch summary", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, summaryResponse{
		Month:            formatMonth(period),
		TotalAmountCents: total,
	})
}

type subscriptionResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        string    `json:"user_id"`
	Plan          string    `json:"plan"`
	AmountCents   int64     `json:"amount_cents"`
	Currency      string    `json:"currency"`
	BillingPeriod string    `json:"billing_period"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (h *SubscriptionHandler) mapModelToResponse(sub *subscription.Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:            sub.ID,
		UserID:        sub.UserID,
		Plan:          sub.Plan,
		AmountCents:   sub.AmountCents,
		Currency:      sub.Currency,
		BillingPeriod: formatMonth(sub.BillingPeriod),
		CreatedAt:     sub.CreatedAt,
		UpdatedAt:     sub.UpdatedAt,
	}
}

func (h *SubscriptionHandler) mapRequestToModel(req createSubscriptionRequest) (*subscription.Subscription, error) {
	if strings.TrimSpace(req.UserID) == "" {
		return nil, errors.New("user_id is required")
	}
	if strings.TrimSpace(req.Plan) == "" {
		return nil, errors.New("plan is required")
	}
	if req.AmountCents < 0 {
		return nil, errors.New("amount_cents must be non-negative")
	}
	if strings.TrimSpace(req.Currency) == "" {
		return nil, errors.New("currency is required")
	}

	period, err := parseMonth(req.BillingPeriod)
	if err != nil {
		return nil, errors.New("billing_period must be in MM-YYYY or YYYY-MM format")
	}

	return &subscription.Subscription{
		UserID:        strings.TrimSpace(req.UserID),
		Plan:          strings.TrimSpace(req.Plan),
		AmountCents:   req.AmountCents,
		Currency:      strings.ToUpper(strings.TrimSpace(req.Currency)),
		BillingPeriod: period,
	}, nil
}

func parseMonth(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("month is required")
	}

	layouts := []string{"01-2006", "2006-01"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, errors.New("invalid month format")
}

func formatMonth(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
