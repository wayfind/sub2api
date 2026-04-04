package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// ---- 错误定义 ----

var (
	ErrWechatPayOrderNotFound   = infraerrors.NotFound("WECHAT_PAY_ORDER_NOT_FOUND", "order not found")
	ErrWechatPayNotEnabled      = infraerrors.BadRequest("WECHAT_PAY_NOT_ENABLED", "wechat pay is not enabled")
	ErrWechatPayNotConfigured   = infraerrors.BadRequest("WECHAT_PAY_NOT_CONFIGURED", "wechat pay is not configured")
	ErrWechatPayInvalidPackage  = infraerrors.BadRequest("WECHAT_PAY_INVALID_PACKAGE", "invalid payment package")
	ErrWechatPayOrderExpired    = infraerrors.BadRequest("WECHAT_PAY_ORDER_EXPIRED", "order has expired")
)

// ---- Setting 键名常量 ----

const (
	SettingKeyWechatPayConfig   = "wechat_pay_config"
	SettingKeyWechatPayPackages = "wechat_pay_packages"
	SettingKeyWechatPayEnabled  = "wechat_pay_enabled"
)

// ---- 数据模型 ----

// WechatPayOrder 微信支付订单
type WechatPayOrder struct {
	ID            int64
	OrderNo       string
	UserID        int64
	PackageID     int
	CnyFee        int     // 人民币金额（分）
	UsdAmount     float64 // 到账美元
	Status        string  // pending / paid / expired / refunded
	WechatTradeNo *string
	CodeURL       *string
	ExpiresAt     time.Time
	PaidAt        *time.Time
	NotifyData    *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// WechatPayConfig 微信支付配置（存 Setting 表）
type WechatPayConfig struct {
	AppID      string `json:"appid"`
	MchID      string `json:"mchid"`
	APIKeyV3   string `json:"api_key_v3"`
	SerialNo   string `json:"serial_no"`
	PrivateKey string `json:"private_key"` // PEM 格式私钥内容
	NotifyURL  string `json:"notify_url"`  // 回调地址
}

// WechatPayPackage 充值套餐（存 Setting 表）
type WechatPayPackage struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	CnyAmount float64 `json:"cny_amount"` // 人民币金额（元）
	UsdAmount float64 `json:"usd_amount"` // 到账美元
}

// ---- Repository 接口 ----

type WechatPayOrderRepository interface {
	Create(ctx context.Context, order *WechatPayOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*WechatPayOrder, error)
	GetByID(ctx context.Context, id int64) (*WechatPayOrder, error)
	// MarkPaid 幂等标记支付成功，返回 true 表示本次更新生效
	MarkPaid(ctx context.Context, orderNo, wechatTradeNo, notifyData string) (bool, error)
	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WechatPayOrder, *pagination.PaginationResult, error)
	List(ctx context.Context, params pagination.PaginationParams, status string) ([]WechatPayOrder, *pagination.PaginationResult, error)
}

// ---- Service ----

type WechatPayService struct {
	db          *dbent.Client
	orderRepo   WechatPayOrderRepository
	settingRepo SettingRepository
	userService *UserService
}

func NewWechatPayService(
	db *dbent.Client,
	orderRepo WechatPayOrderRepository,
	settingRepo SettingRepository,
	userService *UserService,
) *WechatPayService {
	return &WechatPayService{
		db:          db,
		orderRepo:   orderRepo,
		settingRepo: settingRepo,
		userService: userService,
	}
}

// GetConfig 获取微信支付配置
func (s *WechatPayService) GetConfig(ctx context.Context) (*WechatPayConfig, error) {
	val, err := s.settingRepo.GetValue(ctx, SettingKeyWechatPayConfig)
	if err != nil {
		return nil, ErrWechatPayNotConfigured
	}
	var cfg WechatPayConfig
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		return nil, ErrWechatPayNotConfigured
	}
	return &cfg, nil
}

// SaveConfig 保存微信支付配置
func (s *WechatPayService) SaveConfig(ctx context.Context, cfg *WechatPayConfig) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyWechatPayConfig, string(b))
}

// UpdateConfig 更新微信支付配置，私钥和 APIKeyV3 为空时保留已存储的值。
// 前端无法读取这两个字段，更新其他字段时不会传它们，此处在 service 层做 merge。
func (s *WechatPayService) UpdateConfig(ctx context.Context, incoming *WechatPayConfig) error {
	if incoming.PrivateKey == "" || incoming.APIKeyV3 == "" {
		existing, err := s.GetConfig(ctx)
		if err == nil {
			if incoming.PrivateKey == "" {
				incoming.PrivateKey = existing.PrivateKey
			}
			if incoming.APIKeyV3 == "" {
				incoming.APIKeyV3 = existing.APIKeyV3
			}
		}
	}
	return s.SaveConfig(ctx, incoming)
}

// IsEnabled 是否启用微信支付
func (s *WechatPayService) IsEnabled(ctx context.Context) bool {
	val, err := s.settingRepo.GetValue(ctx, SettingKeyWechatPayEnabled)
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(val)) == "true"
}

// SetEnabled 启用/禁用微信支付
func (s *WechatPayService) SetEnabled(ctx context.Context, enabled bool) error {
	v := "false"
	if enabled {
		v = "true"
	}
	return s.settingRepo.Set(ctx, SettingKeyWechatPayEnabled, v)
}

// GetPackages 获取充值套餐列表
func (s *WechatPayService) GetPackages(ctx context.Context) ([]WechatPayPackage, error) {
	val, err := s.settingRepo.GetValue(ctx, SettingKeyWechatPayPackages)
	if err != nil || val == "" {
		return []WechatPayPackage{}, nil
	}
	var pkgs []WechatPayPackage
	if err := json.Unmarshal([]byte(val), &pkgs); err != nil {
		return nil, fmt.Errorf("unmarshal packages: %w", err)
	}
	return pkgs, nil
}

// SavePackages 保存充值套餐列表
func (s *WechatPayService) SavePackages(ctx context.Context, pkgs []WechatPayPackage) error {
	b, err := json.Marshal(pkgs)
	if err != nil {
		return fmt.Errorf("marshal packages: %w", err)
	}
	return s.settingRepo.Set(ctx, SettingKeyWechatPayPackages, string(b))
}

// CreateOrder 创建微信支付订单，返回二维码链接
func (s *WechatPayService) CreateOrder(ctx context.Context, userID int64, packageID int) (*WechatPayOrder, error) {
	if !s.IsEnabled(ctx) {
		return nil, ErrWechatPayNotEnabled
	}

	// 获取套餐
	pkgs, err := s.GetPackages(ctx)
	if err != nil {
		return nil, err
	}
	var pkg *WechatPayPackage
	for i := range pkgs {
		if pkgs[i].ID == packageID {
			pkg = &pkgs[i]
			break
		}
	}
	if pkg == nil {
		return nil, ErrWechatPayInvalidPackage
	}

	// 获取微信支付配置
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	// 生成业务订单号
	orderNo, err := generateOrderNo()
	if err != nil {
		return nil, fmt.Errorf("generate order no: %w", err)
	}

	// 构造微信支付客户端
	client, err := buildWechatClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("build wechat client: %w", err)
	}

	// 调用微信 Native 下单接口
	svc := native.NativeApiService{Client: client}
	cnyFee := int(pkg.CnyAmount * 100) // 元 → 分
	resp, _, err := svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(cfg.AppID),
		Mchid:       core.String(cfg.MchID),
		Description: core.String(pkg.Name),
		OutTradeNo:  core.String(orderNo),
		NotifyUrl:   core.String(cfg.NotifyURL),
		Amount: &native.Amount{
			Currency: core.String("CNY"),
			Total:    core.Int64(int64(cnyFee)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("wechat prepay: %w", err)
	}

	codeURL := ""
	if resp.CodeUrl != nil {
		codeURL = *resp.CodeUrl
	}

	// 保存订单
	order := &WechatPayOrder{
		OrderNo:   orderNo,
		UserID:    userID,
		PackageID: packageID,
		CnyFee:    cnyFee,
		UsdAmount: pkg.UsdAmount,
		Status:    "pending",
		CodeURL:   &codeURL,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return order, nil
}

// GetOrderStatus 查询订单状态（前端轮询用）
func (s *WechatPayService) GetOrderStatus(ctx context.Context, userID int64, orderNo string) (*WechatPayOrder, error) {
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	// 只能查询自己的订单
	if order.UserID != userID {
		return nil, ErrWechatPayOrderNotFound
	}
	// 检查是否过期
	if order.Status == "pending" && time.Now().After(order.ExpiresAt) {
		order.Status = "expired"
	}
	return order, nil
}

// HandleNotify 处理微信支付回调，返回是否为首次成功处理
func (s *WechatPayService) HandleNotify(ctx context.Context, body []byte, cfg *WechatPayConfig) (bool, error) {
	// 解析回调数据，验证签名由调用方（handler）使用 SDK 的 NotifyHandler 完成
	// 这里处理业务逻辑
	var notifyMsg struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
		TradeState    string `json:"trade_state"`
	}
	if err := json.Unmarshal(body, &notifyMsg); err != nil {
		return false, fmt.Errorf("unmarshal notify: %w", err)
	}

	if notifyMsg.TradeState != "SUCCESS" {
		// 非成功状态，返回 true 让微信停止重推
		return false, nil
	}

	// 在事务中原子完成：标记订单已支付 + 给用户加余额
	// 两步必须同时成功，否则回滚，避免"已付款但余额未到账"的不一致状态
	var firstTime bool
	tx, err := s.db.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	defer func() { _ = tx.Rollback() }()

	updated, err := s.orderRepo.MarkPaid(txCtx, notifyMsg.OutTradeNo, notifyMsg.TransactionID, string(body))
	if err != nil {
		return false, fmt.Errorf("mark paid: %w", err)
	}
	if !updated {
		// 已处理过（UPDATE 匹配到 0 行），事务内无有效写入，defer Rollback 安全退出
		return false, nil
	}

	// 查询订单获取用户ID和到账金额
	order, err := s.orderRepo.GetByOrderNo(txCtx, notifyMsg.OutTradeNo)
	if err != nil {
		return false, fmt.Errorf("get order: %w", err)
	}

	// 给用户增加余额
	if err := s.userService.UpdateBalance(txCtx, order.UserID, order.UsdAmount); err != nil {
		return false, fmt.Errorf("update balance: user_id=%d amount=%f err=%w", order.UserID, order.UsdAmount, err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}

	firstTime = true
	log.Printf("wechat pay: order paid: order_no=%s user_id=%d usd_amount=%f",
		order.OrderNo, order.UserID, order.UsdAmount)

	return firstTime, nil
}

// ListOrdersByUser 获取用户充值记录
func (s *WechatPayService) ListOrdersByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]WechatPayOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListByUser(ctx, userID, params)
}

// ListOrders 管理员查询订单列表
func (s *WechatPayService) ListOrders(ctx context.Context, params pagination.PaginationParams, status string) ([]WechatPayOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.List(ctx, params, status)
}

// ---- 工具函数 ----

func generateOrderNo() (string, error) {
	const chars = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b[i] = chars[n.Int64()]
	}
	return fmt.Sprintf("WX%d%s", time.Now().UnixMilli(), string(b)), nil
}

func buildWechatClient(cfg *WechatPayConfig) (*core.Client, error) {
	privateKey, err := utils.LoadPrivateKey(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("load private key: %w", err)
	}
	ctx := context.Background()
	client, err := core.NewClient(ctx,
		option.WithWechatPayAutoAuthCipher(cfg.MchID, cfg.SerialNo, privateKey, cfg.APIKeyV3),
	)
	if err != nil {
		return nil, fmt.Errorf("new wechat client: %w", err)
	}
	return client, nil
}
