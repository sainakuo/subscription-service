package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sainakuo/subscription-service/internal/model"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	db *pgxpool.Pool
}

type SubscriptionFilter struct {
	UserID      *uuid.UUID
	ServiceName string
	Limit       int
	Offset      int
}

func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{
		db: db,
	}
}

func (r *SubscriptionRepository) Create(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error) {
	query := `
		INSERT INTO subscriptions (
			service_name,
			price,
			user_id,
			start_date,
			end_date
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING 
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
	`

	createdSubscription, err := scanSubscription(
		r.db.QueryRow(
			ctx,
			query,
			subscription.ServiceName,
			subscription.Price,
			subscription.UserID,
			subscription.StartDate,
			subscription.EndDate,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return createdSubscription, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	query := `
		SELECT 
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		WHERE id = $1
	`

	subscription, err := scanSubscription(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}

		return nil, fmt.Errorf("failed to get subscription by id: %w", err)
	}

	return subscription, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter SubscriptionFilter) ([]model.Subscription, error) {
	query := `
		SELECT 
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
		FROM subscriptions
		WHERE 1 = 1
	`

	args := make([]any, 0)
	argNumber := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argNumber)
		args = append(args, *filter.UserID)
		argNumber++
	}

	if filter.ServiceName != "" {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argNumber)
		args = append(args, "%"+filter.ServiceName+"%")
		argNumber++
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Offset < 0 {
		filter.Offset = 0
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argNumber, argNumber+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	subscriptions := make([]model.Subscription, 0)

	for rows.Next() {
		subscription, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		subscriptions = append(subscriptions, *subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error while listing subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, subscription *model.Subscription) (*model.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET
			service_name = $2,
			price = $3,
			user_id = $4,
			start_date = $5,
			end_date = $6,
			updated_at = NOW()
		WHERE id = $1
		RETURNING 
			id,
			service_name,
			price,
			user_id,
			start_date,
			end_date,
			created_at,
			updated_at
	`

	updatedSubscription, err := scanSubscription(
		r.db.QueryRow(
			ctx,
			query,
			subscription.ID,
			subscription.ServiceName,
			subscription.Price,
			subscription.UserID,
			subscription.StartDate,
			subscription.EndDate,
		),
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}

		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return updatedSubscription, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM subscriptions
		WHERE id = $1
	`

	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return ErrSubscriptionNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row rowScanner) (*model.Subscription, error) {
	var subscription model.Subscription
	var endDate pgtype.Date

	err := row.Scan(
		&subscription.ID,
		&subscription.ServiceName,
		&subscription.Price,
		&subscription.UserID,
		&subscription.StartDate,
		&endDate,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if endDate.Valid {
		date := endDate.Time
		subscription.EndDate = &date
	}

	return &subscription, nil
}
