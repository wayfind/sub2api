package handler

import (
	"io"
	"log"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
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

	// 用 SDK 的 NotifyHandler 验签并解密回调数据
	privateKey, err := utils.LoadPrivateKey(cfg.PrivateKey)
	if err != nil {
		log.Printf("wechat pay notify: load private key failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "config error"})
		return
	}

	// 使用 downloader 管理器中的证书访问器构建 verifier（复用 CreateOrder 时已注册的 downloader）
	certVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
	verifier := verifiers.NewSHA256WithRSAVerifier(certVisitor)

	handler, err := notify.NewRSANotifyHandler(cfg.APIKeyV3, verifier)
	if err != nil {
		log.Printf("wechat pay notify: new notify handler failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": "FAIL", "message": "handler error"})
		return
	}
	_ = privateKey // 已用于 verifier 构建（通过 downloader 注册）

	// 读取并验签 + 解密
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "read body failed"})
		return
	}

	var transaction map[string]interface{}
	_, err = handler.ParseNotifyRequest(ctx, c.Request, &transaction)
	if err != nil {
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
