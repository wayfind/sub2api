package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/gin-gonic/gin"
)

// RegisterPricingRoutes 注册公开价格查询路由（无需认证）
func RegisterPricingRoutes(v1 *gin.RouterGroup, h *handler.Handlers) {
	pricing := v1.Group("/pricing")
	{
		pricing.GET("/models", h.Pricing.GetPublicModelPricing)
	}
}
