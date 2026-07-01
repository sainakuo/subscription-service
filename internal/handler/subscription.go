package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/model"
	"github.com/sainakuo/subscription-service/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(service *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
	}
}

type CreateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       *int   `json:"price" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date,omitempty"`
}

type UpdateSubscriptionRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	Price       *int   `json:"price" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          string    `json:"id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	UserID      string    `json:"user_id"`
	StartDate   string    `json:"start_date"`
	EndDate     string    `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *SubscriptionHandler) Create(c *gin.Context) {
	var request CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	subscription, err := h.service.Create(
		c.Request.Context(),
		service.CreateSubscriptionInput{
			ServiceName: request.ServiceName,
			Price:       *request.Price,
			UserID:      request.UserID,
			StartDate:   request.StartDate,
			EndDate:     request.EndDate,
		},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, toSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	subscription, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) List(c *gin.Context) {
	limit, err := parseQueryInt(c, "limit", 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "limit must be a valid integer",
		})
		return
	}

	offset, err := parseQueryInt(c, "offset", 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "offset must be a valid integer",
		})
		return
	}

	subscriptions, err := h.service.List(
		c.Request.Context(),
		service.ListSubscriptionsInput{
			UserID:      c.Query("user_id"),
			ServiceName: c.Query("service_name"),
			Limit:       limit,
			Offset:      offset,
		},
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	items := make([]SubscriptionResponse, 0, len(subscriptions))
	for i := range subscriptions {
		items = append(items, toSubscriptionResponse(&subscriptions[i]))
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  items,
		"count":  len(items),
		"limit":  limit,
		"offset": offset,
	})
}

func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var request UpdateSubscriptionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	subscription, err := h.service.Update(
		c.Request.Context(),
		service.UpdateSubscriptionInput{
			ID:          id,
			ServiceName: request.ServiceName,
			Price:       *request.Price,
			UserID:      request.UserID,
			StartDate:   request.StartDate,
			EndDate:     request.EndDate,
		},
	)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseQueryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	value := c.Query(key)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}

func toSubscriptionResponse(subscription *model.Subscription) SubscriptionResponse {
	response := SubscriptionResponse{
		ID:          subscription.ID.String(),
		ServiceName: subscription.ServiceName,
		Price:       subscription.Price,
		UserID:      subscription.UserID.String(),
		StartDate:   service.FormatMonthYear(subscription.StartDate),
		CreatedAt:   subscription.CreatedAt,
		UpdatedAt:   subscription.UpdatedAt,
	}

	if subscription.EndDate != nil {
		response.EndDate = service.FormatMonthYear(*subscription.EndDate)
	}

	return response
}
