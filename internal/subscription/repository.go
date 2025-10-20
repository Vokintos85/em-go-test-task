package subscription

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository provides access to subscription storage.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a Repository backed by pgxpool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a subscription into the database and updates its ID.
func (r *Repository) Create(ctx context.Context, sub *Subscription) error {
	const query = `
        INSERT INTO subscriptions (
            user_id, plan, amount_cents, currency, billing_period
        ) VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at
    `
	row := r.pool.QueryRow(ctx, query,
		sub.UserID,
		sub.Plan,
		sub.AmountCents,
		sub.Currency,
		sub.BillingPeriod,
	)
	return row.Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

// Get retrieves a subscription by its ID.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Subscription, error) {
	const query = `
        SELECT id, user_id, plan, amount_cents, currency, billing_period, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `
	sub := &Subscription{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.UserID,
		&sub.Plan,
		&sub.AmountCents,
		&sub.Currency,
		&sub.BillingPeriod,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return sub, nil
}

// List returns all subscriptions ordered by creation time descending.
func (r *Repository) List(ctx context.Context) ([]Subscription, error) {
	const query = `
        SELECT id, user_id, plan, amount_cents, currency, billing_period, created_at, updated_at
        FROM subscriptions
        ORDER BY created_at DESC
    `
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Subscription
	for rows.Next() {
		var sub Subscription
		err = rows.Scan(
			&sub.ID,
			&sub.UserID,
			&sub.Plan,
			&sub.AmountCents,
			&sub.Currency,
			&sub.BillingPeriod,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, sub)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return result, nil
}

// Update modifies an existing subscription.
func (r *Repository) Update(ctx context.Context, sub *Subscription) error {
	const query = `
        UPDATE subscriptions
        SET user_id = $1,
            plan = $2,
            amount_cents = $3,
            currency = $4,
            billing_period = $5
        WHERE id = $6
        RETURNING updated_at
    `
	row := r.pool.QueryRow(ctx, query,
		sub.UserID,
		sub.Plan,
		sub.AmountCents,
		sub.Currency,
		sub.BillingPeriod,
		sub.ID,
	)
	if err := row.Scan(&sub.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// Delete removes a subscription by ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `
        DELETE FROM subscriptions
        WHERE id = $1
    `
	ct, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// MonthlySummary returns the sum of amount_cents for a billing period month.
func (r *Repository) MonthlySummary(ctx context.Context, period time.Time) (int64, error) {
	const query = `
        SELECT COALESCE(SUM(amount_cents), 0)
        FROM subscriptions
        WHERE billing_period = $1
    `
	var total int64
	err := r.pool.QueryRow(ctx, query, period).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

// ErrNotFound indicates no rows were returned.
var ErrNotFound = errors.New("subscription not found")
