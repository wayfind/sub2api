package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

type WechatPayHandler struct {
	wechatPayService *service.WechatPayService
}

func NewWechatPayHandler(wechatPayService *service.WechatPayService) *WechatPayHandler {
	return &WechatPayHandler{wechatPayService: wechatPayService}
}

// GetPackages 获取充值套餐列表
// GET /api/v1/payments/wechat/packages
func (h *WechatPayHandler) GetPackages(c *gin.Context) {
	pkgs, err := h.wechatPayService.GetPackages(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, pkgs)
}

type createOrderRequest struct {
	PackageID int `json:"package_id" binding:"required"`
}

// CreateOrder 创建支付订单，返回二维码链接
// POST /api/v1/payments/wechat/create-order
func (h *WechatPayHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	order, err := h.wechatPayService.CreateOrder(c.Request.Context(), subject.UserID, req.PackageID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	codeURL := ""
	if order.CodeURL != nil {
		codeURL = *order.CodeURL
	}

	response.Success(c, gin.H{
		"order_no": order.OrderNo,
		"code_url": codeURL,
		"expires_at": order.ExpiresAt,
	})
}

// GetOrderStatus 查询订单状态（前端轮询）
// GET /api/v1/payments/wechat/order/:order_no
func (h *WechatPayHandler) GetOrderStatus(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	orderNo := c.Param("order_no")
	if orderNo == "" {
		response.BadRequest(c, "order_no is required")
		return
	}

	order, err := h.wechatPayService.GetOrderStatus(c.Request.Context(), subject.UserID, orderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{
		"order_no":   order.OrderNo,
		"status":     order.Status,
		"cny_fee":    order.CnyFee,
		"usd_amount": order.UsdAmount,
		"paid_at":    order.PaidAt,
		"expires_at": order.ExpiresAt,
	})
}

// HandleNotify 接收微信支付回调
// POST /api/v1/payments/wechat/notify
func (h *WechatPayHandler) HandleNotify(c *gin.Context) {
	ctx := c.Request.Context()

	cfg, err := h.wechatPayService.GetConfig(ctx)
	if err != nil {
		log.Printf("wechat pay notify: get config failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "config error"})
		return
	}

	if cfg.PublicKeyID == "" || cfg.PublicKey == "" {
		log.Printf("wechat pay notify: public key not configured, rejecting callback")
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "public key not configured"})
		return
	}

	publicKey, err := utils.LoadPublicKey(strings.ReplaceAll(cfg.PublicKey, `\n`, "\n"))
	if err != nil {
		log.Printf("wechat pay notify: load public key failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "config error"})
		return
	}

	notifyHandler, err := notify.NewRSANotifyHandler(
		cfg.APIKeyV3,
		verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *publicKey),
	)
	if err != nil {
		log.Printf("wechat pay notify: new notify handler failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "handler error"})
		return
	}

	// ParseNotifyRequest 内部完成：验签 + AES-GCM 解密 + 反序列化明文到 transaction
	// SDK 内部会自行处理 body 的读取和恢复，无需手动操作
	var transaction map[string]interface{}
	if _, err = notifyHandler.ParseNotifyRequest(ctx, c.Request, &transaction); err != nil {
		log.Printf("wechat pay notify: parse request failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid signature"})
		return
	}

	// 将解密后的明文业务数据传给 service，而非原始加密 body
	plaintext, err := json.Marshal(transaction)
	if err != nil {
		log.Printf("wechat pay notify: marshal transaction failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "internal error"})
		return
	}

	if _, err := h.wechatPayService.HandleNotify(ctx, plaintext); err != nil {
		log.Printf("wechat pay notify: handle notify failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}

// GetOrders 获取用户充值记录
// GET /api/v1/payments/wechat/orders
func (h *WechatPayHandler) GetOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, pageSize := response.ParsePagination(c)
	orders, result, err := h.wechatPayService.ListOrdersByUser(
		c.Request.Context(), subject.UserID,
		pagination.PaginationParams{Page: page, PageSize: pageSize},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, orders, result.Total, page, pageSize)
}
