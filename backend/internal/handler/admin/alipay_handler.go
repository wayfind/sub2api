package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// AlipayHandler handles admin alipay management
type AlipayHandler struct {
	alipayService *service.AlipayService
}

func NewAlipayHandler(alipayService *service.AlipayService) *AlipayHandler {
	return &AlipayHandler{alipayService: alipayService}
}

// GetConfig 获取支付宝配置（屏蔽敏感字段）
// GET /api/v1/admin/alipay/config
func (h *AlipayHandler) GetConfig(c *gin.Context) {
	cfg, err := h.alipayService.GetConfig(c.Request.Context())
	notifyURL, _ := h.alipayService.NotifyURL(c.Request.Context())
	if err != nil {
		response.Success(c, gin.H{
			"notify_url": notifyURL,
		})
		return
	}
	response.Success(c, gin.H{
		"app_id":          cfg.AppID,
		"notify_url":      notifyURL,
		"is_prod":         cfg.IsProd,
		"private_key_set": cfg.PrivateKey != "",
		"public_key_set":  cfg.PublicKey != "",
		"configured":      true,
	})
}

type alipayUpdateConfigRequest struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	IsProd     bool   `json:"is_prod"`
}

// UpdateConfig 更新支付宝配置
// PUT /api/v1/admin/alipay/config
func (h *AlipayHandler) UpdateConfig(c *gin.Context) {
	var req alipayUpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	// 统一换行符
	req.PrivateKey = strings.ReplaceAll(req.PrivateKey, `\n`, "\n")
	req.PublicKey = strings.ReplaceAll(req.PublicKey, `\n`, "\n")

	cfg := &service.AlipayConfig{
		AppID:      req.AppID,
		PrivateKey: req.PrivateKey,
		PublicKey:  req.PublicKey,
		IsProd:     req.IsProd,
	}
	if err := h.alipayService.UpdateConfig(c.Request.Context(), cfg); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

type alipayUpdateEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// SetEnabled 启用/禁用支付宝支付
// PUT /api/v1/admin/alipay/enabled
func (h *AlipayHandler) SetEnabled(c *gin.Context) {
	var req alipayUpdateEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.alipayService.SetEnabled(c.Request.Context(), req.Enabled); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

// ListOrders 查询订单列表
// GET /api/v1/admin/alipay/orders
func (h *AlipayHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	orders, result, err := h.alipayService.ListOrders(
		c.Request.Context(),
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		status,
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, orders, result.Total, page, pageSize)
}
