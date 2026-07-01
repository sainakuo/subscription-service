package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sainakuo/subscription-service/internal/model"
	"github.com/sainakuo/subscription-service/internal/repository"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

var ErrSubscriptionNotFound = errors.New("subscription not found")

type CreateSubscriptionInput struct {
	ServiceName string
	Price       int
	UserID      string
	StartDate   string
	EndDate     string
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, input CreateSubscriptionInput) (*model.Subscription, error) {
	serviceName := strings.TrimSpace(input.ServiceName)
	if serviceName == "" {
		return nil, fmt.Errorf("service_name is required")
	}

	if input.Price <= 0 {
		return nil, fmt.Errorf("price must be greater than 0")
	}

	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return nil, fmt.Errorf("user_id must be a valid UUID")
	}

	startDate, err := ParseMonthYear(input.StartDate)
	if err != nil {
		return nil, fmt.Errorf("start_date must be in MM-YYYY format")
	}

	var endDate *time.Time
	if strings.TrimSpace(input.EndDate) != "" {
		parsedEndDate, err := ParseMonthYear(input.EndDate)
		if err != nil {
			return nil, fmt.Errorf("end_date must be in MM-YYYY format")
		}

		if parsedEndDate.Before(startDate) {
			return nil, fmt.Errorf("end_date cannot be before start_date")
		}

		endDate = &parsedEndDate
	}

	subscription := &model.Subscription{
		ServiceName: serviceName,
		Price:       input.Price,
		UserID:      userID,
		StartDate:   startDate,
		EndDate:     endDate,
	}

	return s.repo.Create(ctx, subscription)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id string) (*model.Subscription, error) {
	subscriptionID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("id must be a valid UUID")
	}

	subscription, err := s.repo.GetByID(ctx, subscriptionID)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriptionNotFound) {
			return nil, ErrSubscriptionNotFound
		}

		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	return subscription, nil
}

func ParseMonthYear(value string) (time.Time, error) {
	value = strings.TrimSpace(value)

	return time.Parse("02-01-2006", "01-"+value)
}

func FormatMonthYear(value time.Time) string {
	return value.Format("01-2006")
}
