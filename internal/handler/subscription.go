package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sainakuo/subscription-service/internal/model"
	"github.com/sainakuo/subscription-service/internal/service"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
	log     *slog.Logger
}

func NewSubscriptionHandler(service *service.SubscriptionService, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: service,
		log:     log,
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

type TotalCostResponse struct {
	TotalPrice  int    `json:"total_price"`
	Currency    string `json:"currency"`
	From        string `json:"from"`
	To          string `json:"to"`
	UserID      string `json:"user_id,omitempty"`
	ServiceName string `json:"service_name,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ListSubscriptionsResponse struct {
	Items  []SubscriptionResponse `json:"items"`
	Count  int                    `json:"count"`
	Limit  int                    `json:"limit"`
	Offset int                    `json:"offset"`
}

// Create godoc
// @Summary Create subscription
// @Description Create a new user subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param input body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *SubscriptionHandler) Create(c *gin.Context) {

	var request CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.Warn("invalid create subscription request", "error", err)
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
		h.log.Warn("failed to create subscription", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.log.Info("subscription created",
		"subscription_id", subscription.ID.String(),
		"user_id", subscription.UserID.String(),
		"service_name", subscription.ServiceName,
	)

	c.JSON(http.StatusCreated, toSubscriptionResponse(subscription))
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Get a subscription by its UUID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
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

// List godoc
// @Summary List subscriptions
// @Description Get subscriptions list with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service name"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} ListSubscriptionsResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions [get]
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

	c.JSON(http.StatusOK, ListSubscriptionsResponse{
		Items:  items,
		Count:  len(items),
		Limit:  limit,
		Offset: offset,
	})
}

// Update godoc
// @Summary Update subscription
// @Description Update an existing subscription by ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "Subscription ID"
// @Param input body UpdateSubscriptionRequest true "Updated subscription data"
// @Success 200 {object} SubscriptionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
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
		h.log.Warn("failed to update subscription", "id", id, "error", err)
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

	h.log.Info("subscription updated",
		"subscription_id", subscription.ID.String(),
		"user_id", subscription.UserID.String(),
		"service_name", subscription.ServiceName,
	)

	c.JSON(http.StatusOK, toSubscriptionResponse(subscription))
}

// Delete godoc
// @Summary Delete subscription
// @Description Delete a subscription by ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "Subscription ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		h.log.Warn("failed to delete subscription", "id", id, "error", err)
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

	h.log.Info("subscription deleted", "subscription_id", id)

	c.Status(http.StatusNoContent)
}

// CalculateTotalCost godoc
// @Summary Calculate total subscriptions cost
// @Description Calculate total cost of subscriptions for selected period with optional filters
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start period in MM-YYYY format"
// @Param to query string true "End period in MM-YYYY format"
// @Param user_id query string false "User UUID"
// @Param service_name query string false "Service name"
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions/total-cost [get]
func (h *SubscriptionHandler) CalculateTotalCost(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "from is required",
		})
		return
	}

	if to == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "to is required",
		})
		return
	}

	result, err := h.service.CalculateTotalCost(
		c.Request.Context(),
		service.TotalCostInput{
			From:        from,
			To:          to,
			UserID:      c.Query("user_id"),
			ServiceName: c.Query("service_name"),
		},
	)
	if err != nil {
		h.log.Warn("failed to calculate total cost", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.log.Info("total cost calculated",
		"total_price", result.TotalPrice,
		"currency", result.Currency,
		"from", result.From,
		"to", result.To,
		"user_id", result.UserID,
		"service_name", result.ServiceName,
	)

	c.JSON(http.StatusOK, TotalCostResponse{
		TotalPrice:  result.TotalPrice,
		Currency:    result.Currency,
		From:        result.From,
		To:          result.To,
		UserID:      result.UserID,
		ServiceName: result.ServiceName,
	})
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
