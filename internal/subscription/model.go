package subscription

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents a persisted subscription record.
type Subscription struct {
	ID            uuid.UUID
	UserID        string
	Plan          string
	AmountCents   int64
	Currency      string
	BillingPeriod time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
