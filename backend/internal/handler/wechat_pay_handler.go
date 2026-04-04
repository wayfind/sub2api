package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"

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

	// 获取微信支付配置（用于验签）
	cfg, err := h.wechatPayService.GetConfig(ctx)
	if err != nil {
		log.Printf("wechat pay notify: get config failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "config error"})
		return
	}

	// 读取请求体（ParseNotifyRequest 内部也会读，先保存一份用于业务处理）
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "read body failed"})
		return
	}
	// 恢复 body 供 SDK 再次读取
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	// 构造验签器：公钥模式（新商户号）或证书模式（旧商户号）
	var notifyHandler *notify.Handler
	if cfg.PublicKeyID != "" && cfg.PublicKey != "" {
		// 新商户号：使用微信支付公钥验签
		publicKey, err := utils.LoadPublicKey(cfg.PublicKey)
		if err != nil {
			log.Printf("wechat pay notify: load public key failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "config error"})
			return
		}
		v := verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *publicKey)
		notifyHandler, err = notify.NewRSANotifyHandler(cfg.APIKeyV3, v)
		if err != nil {
			log.Printf("wechat pay notify: new notify handler failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "handler error"})
			return
		}
	} else {
		// 公钥未配置，无法验签，拒绝处理
		log.Printf("wechat pay notify: public key not configured, rejecting callback")
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "public key not configured"})
		return
	}

	// 验签 + 解密
	var transaction map[string]interface{}
	if _, err = notifyHandler.ParseNotifyRequest(ctx, c.Request, &transaction); err != nil {
		log.Printf("wechat pay notify: parse request failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "invalid signature"})
		return
	}

	// 业务处理
	if _, err := h.wechatPayService.HandleNotify(ctx, bodyBytes, cfg); err != nil {
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
