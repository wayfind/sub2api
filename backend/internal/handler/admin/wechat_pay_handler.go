package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// WechatPayHandler handles admin wechat pay management
type WechatPayHandler struct {
	wechatPayService *service.WechatPayService
}

func NewWechatPayHandler(wechatPayService *service.WechatPayService) *WechatPayHandler {
	return &WechatPayHandler{wechatPayService: wechatPayService}
}

// GetConfig 获取微信支付配置（屏蔽敏感字段）
// GET /api/v1/admin/wechat-pay/config
func (h *WechatPayHandler) GetConfig(c *gin.Context) {
	cfg, err := h.wechatPayService.GetConfig(c.Request.Context())
	if err != nil {
		// 未配置时返回空对象（含自动生成的 notify_url 供参考）
		response.Success(c, gin.H{
			"notify_url": h.wechatPayService.NotifyURL(),
		})
		return
	}
	// 私钥和 APIv3 Key 不返回给前端，只告知是否已配置
	response.Success(c, gin.H{
		"appid":           cfg.AppID,
		"mchid":           cfg.MchID,
		"serial_no":       cfg.SerialNo,
		"notify_url":      h.wechatPayService.NotifyURL(), // 系统自动生成，只读
		"private_key_set": cfg.PrivateKey != "",
		"api_key_v3_set":  cfg.APIKeyV3 != "",
		"configured":      true,
	})
}

type updateConfigRequest struct {
	AppID      string `json:"appid"`
	MchID      string `json:"mchid"`
	APIKeyV3   string `json:"api_key_v3"`
	SerialNo   string `json:"serial_no"`
	PrivateKey string `json:"private_key"`
}

// UpdateConfig 更新微信支付配置
// PUT /api/v1/admin/wechat-pay/config
func (h *WechatPayHandler) UpdateConfig(c *gin.Context) {
	var req updateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	cfg := &service.WechatPayConfig{
		AppID:      req.AppID,
		MchID:      req.MchID,
		APIKeyV3:   req.APIKeyV3,
		SerialNo:   req.SerialNo,
		PrivateKey: req.PrivateKey,
	}
	if err := h.wechatPayService.UpdateConfig(c.Request.Context(), cfg); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

type updateEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// SetEnabled 启用/禁用微信支付
// PUT /api/v1/admin/wechat-pay/enabled
func (h *WechatPayHandler) SetEnabled(c *gin.Context) {
	var req updateEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.wechatPayService.SetEnabled(c.Request.Context(), req.Enabled); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

// GetPackages 获取充值套餐列表
// GET /api/v1/admin/wechat-pay/packages
func (h *WechatPayHandler) GetPackages(c *gin.Context) {
	pkgs, err := h.wechatPayService.GetPackages(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, pkgs)
}

// UpdatePackages 保存充值套餐列表
// PUT /api/v1/admin/wechat-pay/packages
func (h *WechatPayHandler) UpdatePackages(c *gin.Context) {
	var pkgs []service.WechatPayPackage
	if err := c.ShouldBindJSON(&pkgs); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if err := h.wechatPayService.SavePackages(c.Request.Context(), pkgs); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

// ListOrders 查询订单列表
// GET /api/v1/admin/wechat-pay/orders
func (h *WechatPayHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	status := c.Query("status")

	orders, result, err := h.wechatPayService.ListOrders(
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
