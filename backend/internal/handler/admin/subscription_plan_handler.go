package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanHandler handles admin subscription plan management
type SubscriptionPlanHandler struct {
	planService *service.SubscriptionPlanService
}

// NewSubscriptionPlanHandler creates a new admin subscription plan handler
func NewSubscriptionPlanHandler(planService *service.SubscriptionPlanService) *SubscriptionPlanHandler {
	return &SubscriptionPlanHandler{planService: planService}
}

// CreateSubscriptionPlanRequest represents create subscription plan request
type CreateSubscriptionPlanRequest struct {
	Name                string   `json:"name" binding:"required"`
	Description         string   `json:"description"`
	Visibility          string   `json:"visibility"`
	DailyLimitUSD       *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD      *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD     *float64 `json:"monthly_limit_usd"`
	DefaultValidityDays int      `json:"default_validity_days"`
	Price               *float64 `json:"price"`
	SortOrder           int      `json:"sort_order"`
}

// UpdateSubscriptionPlanRequest represents update subscription plan request
type UpdateSubscriptionPlanRequest struct {
	Name                *string  `json:"name"`
	Description         *string  `json:"description"`
	Visibility          *string  `json:"visibility"`
	Status              *string  `json:"status"`
	DailyLimitUSD       *float64 `json:"daily_limit_usd"`
	WeeklyLimitUSD      *float64 `json:"weekly_limit_usd"`
	MonthlyLimitUSD     *float64 `json:"monthly_limit_usd"`
	DefaultValidityDays *int     `json:"default_validity_days"`
	Price               *float64 `json:"price"`
	SortOrder           *int     `json:"sort_order"`
}

// List handles listing subscription plans with pagination
// GET /api/v1/admin/subscription-plans
func (h *SubscriptionPlanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	visibility := c.Query("visibility")
	status := c.Query("status")

	plans, pagination, err := h.planService.List(c.Request.Context(), page, pageSize, visibility, status)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.SubscriptionPlan, 0, len(plans))
	for i := range plans {
		out = append(out, dto.SubscriptionPlanFromService(&plans[i]))
	}
	response.PaginatedWithResult(c, out, toResponsePagination(pagination))
}

// ListAll handles listing all active subscription plans (no pagination)
// GET /api/v1/admin/subscription-plans/all
func (h *SubscriptionPlanHandler) ListAll(c *gin.Context) {
	plans, err := h.planService.ListAll(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	out := make([]*dto.SubscriptionPlan, 0, len(plans))
	for i := range plans {
		out = append(out, dto.SubscriptionPlanFromService(&plans[i]))
	}
	response.Success(c, out)
}

// GetByID handles getting a subscription plan by ID
// GET /api/v1/admin/subscription-plans/:id
func (h *SubscriptionPlanHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	plan, err := h.planService.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.SubscriptionPlanFromService(plan))
}

// Create handles creating a subscription plan
// POST /api/v1/admin/subscription-plans
func (h *SubscriptionPlanHandler) Create(c *gin.Context) {
	var req CreateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.Create(c.Request.Context(), &service.CreateSubscriptionPlanInput{
		Name:                req.Name,
		Description:         req.Description,
		Visibility:          req.Visibility,
		DailyLimitUSD:       req.DailyLimitUSD,
		WeeklyLimitUSD:      req.WeeklyLimitUSD,
		MonthlyLimitUSD:     req.MonthlyLimitUSD,
		DefaultValidityDays: req.DefaultValidityDays,
		Price:               req.Price,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.SubscriptionPlanFromService(plan))
}

// Update handles updating a subscription plan
// PUT /api/v1/admin/subscription-plans/:id
func (h *SubscriptionPlanHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	var req UpdateSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	plan, err := h.planService.Update(c.Request.Context(), id, &service.UpdateSubscriptionPlanInput{
		Name:                req.Name,
		Description:         req.Description,
		Visibility:          req.Visibility,
		Status:              req.Status,
		DailyLimitUSD:       req.DailyLimitUSD,
		WeeklyLimitUSD:      req.WeeklyLimitUSD,
		MonthlyLimitUSD:     req.MonthlyLimitUSD,
		DefaultValidityDays: req.DefaultValidityDays,
		Price:               req.Price,
		SortOrder:           req.SortOrder,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.SubscriptionPlanFromService(plan))
}

// Delete handles soft-deleting a subscription plan
// DELETE /api/v1/admin/subscription-plans/:id
func (h *SubscriptionPlanHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	if err := h.planService.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Subscription plan deleted successfully"})
}
