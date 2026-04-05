package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
)

// ---- 错误定义 ----

var (
	ErrAlipayOrderNotFound  = infraerrors.NotFound("ALIPAY_ORDER_NOT_FOUND", "order not found")
	ErrAlipayNotEnabled     = infraerrors.BadRequest("ALIPAY_NOT_ENABLED", "alipay is not enabled")
	ErrAlipayNotConfigured  = infraerrors.BadRequest("ALIPAY_NOT_CONFIGURED", "alipay is not configured")
	ErrAlipayInvalidPackage = infraerrors.BadRequest("ALIPAY_INVALID_PACKAGE", "invalid payment package")
)

// ---- Setting 键名常量 ----

const (
	SettingKeyAlipayConfig  = "alipay_config"
	SettingKeyAlipayEnabled = "alipay_enabled"
)

// ---- 数据模型 ----

// AlipayOrder 支付宝订单
type AlipayOrder struct {
	ID            int64
	OrderNo       string
	UserID        int64
	PackageID     int
	CnyFee        int     // 人民币金额（分）
	UsdAmount     float64 // 到账美元
	Status        string  // pending / paid / expired / refunded
	AlipayTradeNo *string
	QRCode        *string
	ExpiresAt     time.Time
	PaidAt        *time.Time
	NotifyData    *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AlipayConfig 支付宝支付配置（存 Setting 表）
type AlipayConfig struct {
	AppID      string `json:"app_id"`
	PrivateKey string `json:"private_key"` // 应用私钥（PKCS1 或 PKCS8 PEM，或裸 base64）
	PublicKey  string `json:"public_key"`  // 支付宝公钥（裸 base64 或 PEM）
	IsProd     bool   `json:"is_prod"`     // true=正式环境，false=沙箱
}

// ---- Repository 接口 ----

type AlipayOrderRepository interface {
	Create(ctx context.Context, order *AlipayOrder) error
	GetByOrderNo(ctx context.Context, orderNo string) (*AlipayOrder, error)
	// MarkPaid 幂等标记支付成功，返回 true 表示本次更新生效
	MarkPaid(ctx context.Context, orderNo, alipayTradeNo, notifyData string) (bool, error)
	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]AlipayOrder, *pagination.PaginationResult, error)
	List(ctx context.Context, params pagination.PaginationParams, status string) ([]AlipayOrder, *pagination.PaginationResult, error)
}

// ---- Service ----

type AlipayService struct {
	db          *dbent.Client
	cfg         *config.Config
	orderRepo   AlipayOrderRepository
	settingRepo SettingRepository
	userService *UserService

	// 缓存已构建的 alipay client，避免每次创建订单都重新解析私钥。
	// cacheKey = sha256(appID + privateKey + isProd)，任何字段变更都会重建。
	clientMu     sync.Mutex
	clientCacheKey string
	cachedClient *alipay.Client
}

func NewAlipayService(
	db *dbent.Client,
	cfg *config.Config,
	orderRepo AlipayOrderRepository,
	settingRepo SettingRepository,
	userService *UserService,
) *AlipayService {
	return &AlipayService{
		db:          db,
		cfg:         cfg,
		orderRepo:   orderRepo,
		settingRepo: settingRepo,
		userService: userService,
	}
}

// NotifyURL 生成支付宝回调地址
func (s *AlipayService) NotifyURL(ctx context.Context) (string, bool) {
	base := s.cfg.Server.FrontendURL
	if val, err := s.settingRepo.GetValue(ctx, SettingKeyFrontendURL); err == nil && strings.TrimSpace(val) != "" {
		base = strings.TrimSpace(val)
	}
	base = strings.TrimRight(base, "/")
	u := base + "/api/v1/payments/alipay/notify"
	valid := strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
	return u, valid
}

// GetConfig 获取支付宝配置，优先读 config.yaml，未配置时回落到 Setting 表
func (s *AlipayService) GetConfig(ctx context.Context) (*AlipayConfig, error) {
	// config.yaml 优先：AppID 非空即视为已配置
	if s.cfg.Alipay.AppID != "" {
		if s.cfg.Alipay.PrivateKey == "" || s.cfg.Alipay.PublicKey == "" {
			return nil, fmt.Errorf("alipay config in config.yaml is incomplete: missing private_key or public_key")
		}
		return &AlipayConfig{
			AppID:      s.cfg.Alipay.AppID,
			PrivateKey: s.cfg.Alipay.PrivateKey,
			PublicKey:  s.cfg.Alipay.PublicKey,
			IsProd:     s.cfg.Alipay.IsProd,
		}, nil
	}
	// 回落到 Setting 表（管理后台手动配置）
	val, err := s.settingRepo.GetValue(ctx, SettingKeyAlipayConfig)
	if err != nil {
		return nil, ErrAlipayNotConfigured
	}
	var cfg AlipayConfig
	if err := json.Unmarshal([]byte(val), &cfg); err != nil {
		return nil, ErrAlipayNotConfigured
	}
	return &cfg, nil
}

// getOrBuildClient 返回缓存的 alipay client。
// cache key = sha256(appID|privateKey|isProd)[:8]，任意字段变更都会触发重建。
func (s *AlipayService) getOrBuildClient(cfg *AlipayConfig) (*alipay.Client, error) {
	isProdStr := "0"
	if cfg.IsProd {
		isProdStr = "1"
	}
	h := sha256.Sum256([]byte(cfg.AppID + "|" + cfg.PrivateKey + "|" + isProdStr))
	key := fmt.Sprintf("%x", h[:8])

	s.clientMu.Lock()
	defer s.clientMu.Unlock()
	if s.cachedClient != nil && s.clientCacheKey == key {
		return s.cachedClient, nil
	}
	client, err := buildAlipayClient(cfg)
	if err != nil {
		return nil, err
	}
	s.cachedClient = client
	s.clientCacheKey = key
	return client, nil
}

// SaveConfig 保存支付宝配置，同时失效 client 缓存
func (s *AlipayService) SaveConfig(ctx context.Context, cfg *AlipayConfig) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyAlipayConfig, string(b)); err != nil {
		return err
	}
	// config 变更，清空缓存的 client
	s.clientMu.Lock()
	s.cachedClient = nil
	s.clientCacheKey = ""
	s.clientMu.Unlock()
	return nil
}

// UpdateConfig 更新配置，私钥/公钥为空时保留已存储的值
func (s *AlipayService) UpdateConfig(ctx context.Context, incoming *AlipayConfig) error {
	if incoming.PrivateKey == "" || incoming.PublicKey == "" {
		existing, err := s.GetConfig(ctx)
		if err == nil {
			if incoming.PrivateKey == "" {
				incoming.PrivateKey = existing.PrivateKey
			}
			if incoming.PublicKey == "" {
				incoming.PublicKey = existing.PublicKey
			}
		}
	}
	return s.SaveConfig(ctx, incoming)
}

// IsEnabled 是否启用支付宝支付，config.yaml 优先于 Setting 表
func (s *AlipayService) IsEnabled(ctx context.Context) bool {
	// config.yaml 中 AppID 非空时，以 enabled 字段为准
	if s.cfg.Alipay.AppID != "" {
		return s.cfg.Alipay.Enabled
	}
	// 回落到 Setting 表
	val, err := s.settingRepo.GetValue(ctx, SettingKeyAlipayEnabled)
	if err != nil {
		return false
	}
	return strings.ToLower(strings.TrimSpace(val)) == "true"
}

// SetEnabled 启用/禁用支付宝支付
func (s *AlipayService) SetEnabled(ctx context.Context, enabled bool) error {
	v := "false"
	if enabled {
		v = "true"
	}
	return s.settingRepo.Set(ctx, SettingKeyAlipayEnabled, v)
}

// CreateOrder 创建支付宝当面付订单，返回二维码链接
func (s *AlipayService) CreateOrder(ctx context.Context, userID int64, packageID int) (*AlipayOrder, error) {
	if !s.IsEnabled(ctx) {
		return nil, ErrAlipayNotEnabled
	}

	notifyURL, valid := s.NotifyURL(ctx)
	if !valid {
		return nil, ErrAlipayNotConfigured
	}

	// 复用微信套餐
	pkgs, err := s.wechatPackages(ctx)
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
		return nil, ErrAlipayInvalidPackage
	}

	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	orderNo, err := generateAlipayOrderNo()
	if err != nil {
		return nil, fmt.Errorf("generate order no: %w", err)
	}

	client, err := s.getOrBuildClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("build alipay client: %w", err)
	}

	cnyFee := int(math.Round(pkg.CnyAmount * 100))
	totalAmount := fmt.Sprintf("%d.%02d", cnyFee/100, cnyFee%100)

	bm := make(gopay.BodyMap)
	bm.Set("subject", pkg.Name).
		Set("out_trade_no", orderNo).
		Set("total_amount", totalAmount).
		Set("notify_url", notifyURL)

	resp, err := client.TradePrecreate(ctx, bm)
	if err != nil {
		return nil, fmt.Errorf("alipay precreate: %w", err)
	}
	if resp.Response == nil || resp.Response.QrCode == "" {
		code, msg := "", ""
		if resp.Response != nil {
			code = resp.Response.Code
			msg = resp.Response.Msg
		}
		return nil, fmt.Errorf("alipay precreate: empty qr_code, code=%s msg=%s", code, msg)
	}

	qrCode := resp.Response.QrCode

	order := &AlipayOrder{
		OrderNo:   orderNo,
		UserID:    userID,
		PackageID: packageID,
		CnyFee:    cnyFee,
		UsdAmount: pkg.UsdAmount,
		Status:    "pending",
		QRCode:    &qrCode,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	log.Printf("alipay: order created: order_no=%s user_id=%d package_id=%d cny_fee=%d usd_amount=%.4f",
		order.OrderNo, order.UserID, order.PackageID, order.CnyFee, order.UsdAmount)

	return order, nil
}

// GetOrderStatus 查询订单状态（前端轮询用）
func (s *AlipayService) GetOrderStatus(ctx context.Context, userID int64, orderNo string) (*AlipayOrder, error) {
	order, err := s.orderRepo.GetByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.UserID != userID {
		return nil, ErrAlipayOrderNotFound
	}
	if order.Status == "pending" && time.Now().After(order.ExpiresAt) {
		order.Status = "expired"
	}
	return order, nil
}

// HandleNotify 处理支付宝异步通知（已由 handler 层完成验签）
func (s *AlipayService) HandleNotify(ctx context.Context, notifyMap map[string]any) (bool, error) {
	tradeStatus, _ := notifyMap["trade_status"].(string)
	outTradeNo, _ := notifyMap["out_trade_no"].(string)
	alipayTradeNo, _ := notifyMap["trade_no"].(string)

	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		log.Printf("alipay notify: ignored: out_trade_no=%s trade_status=%s", outTradeNo, tradeStatus)
		return false, nil
	}

	if outTradeNo == "" || alipayTradeNo == "" {
		log.Printf("alipay notify: missing required fields: out_trade_no=%q trade_no=%q trade_status=%s", outTradeNo, alipayTradeNo, tradeStatus)
		return false, nil
	}

	notifyJSON, err := json.Marshal(notifyMap)
	if err != nil {
		log.Printf("alipay notify: marshal notify data failed (audit loss): %v", err)
		notifyJSON = []byte("{}")
	}

	tx, err := s.db.Tx(ctx)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	txCtx := dbent.NewTxContext(ctx, tx)
	defer func() { _ = tx.Rollback() }()

	updated, err := s.orderRepo.MarkPaid(txCtx, outTradeNo, alipayTradeNo, string(notifyJSON))
	if err != nil {
		return false, fmt.Errorf("mark paid: %w", err)
	}
	if !updated {
		return false, nil
	}

	order, err := s.orderRepo.GetByOrderNo(txCtx, outTradeNo)
	if err != nil {
		return false, fmt.Errorf("get order: %w", err)
	}

	if err := s.userService.UpdateBalance(txCtx, order.UserID, order.UsdAmount); err != nil {
		return false, fmt.Errorf("update balance: user_id=%d amount=%f err=%w", order.UserID, order.UsdAmount, err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit tx: %w", err)
	}

	log.Printf("alipay: order paid: order_no=%s alipay_trade_no=%s user_id=%d cny_fee=%d usd_amount=%.4f",
		order.OrderNo, alipayTradeNo, order.UserID, order.CnyFee, order.UsdAmount)

	return true, nil
}

// ListOrdersByUser 获取用户充值记录
func (s *AlipayService) ListOrdersByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]AlipayOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.ListByUser(ctx, userID, params)
}

// ListOrders 管理员查询订单列表
func (s *AlipayService) ListOrders(ctx context.Context, params pagination.PaginationParams, status string) ([]AlipayOrder, *pagination.PaginationResult, error) {
	return s.orderRepo.List(ctx, params, status)
}

// GetPackages 返回当前配置的充值套餐列表
func (s *AlipayService) GetPackages(ctx context.Context) ([]WechatPayPackage, error) {
	return s.wechatPackages(ctx)
}

// wechatPackages 复用微信套餐配置
func (s *AlipayService) wechatPackages(ctx context.Context) ([]WechatPayPackage, error) {
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

// ---- 工具函数 ----

func generateAlipayOrderNo() (string, error) {
	const chars = "0123456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	b := make([]byte, 6)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b[i] = chars[n.Int64()]
	}
	return fmt.Sprintf("AP%d%s", time.Now().UnixMilli(), string(b)), nil
}

func buildAlipayClient(cfg *AlipayConfig) (*alipay.Client, error) {
	// 处理字面 \n，剥去 PEM headers，得到裸 base64
	privateKey := stripPEMHeaders(strings.ReplaceAll(cfg.PrivateKey, `\n`, "\n"))

	// gopay 要求 PKCS1 格式的裸 base64；若用户提供的是 PKCS8，自动转换
	privateKey, err := ensurePKCS1(privateKey)
	if err != nil {
		return nil, fmt.Errorf("convert private key to PKCS1: %w", err)
	}

	client, err := alipay.NewClient(cfg.AppID, privateKey, cfg.IsProd)
	if err != nil {
		return nil, fmt.Errorf("new alipay client: %w", err)
	}

	if cfg.PublicKey != "" {
		pubKeyPEM := strings.ReplaceAll(cfg.PublicKey, `\n`, "\n")
		if !strings.Contains(pubKeyPEM, "-----") {
			pubKeyPEM = "-----BEGIN PUBLIC KEY-----\n" + pubKeyPEM + "\n-----END PUBLIC KEY-----\n"
		}
		client.AutoVerifySign([]byte(pubKeyPEM))
	}

	return client, nil
}

// ensurePKCS1 若输入是 PKCS8 裸 base64，转为 PKCS1 裸 base64；PKCS1 直接返回
func ensurePKCS1(bareBase64 string) (string, error) {
	der, err := base64.StdEncoding.DecodeString(bareBase64)
	if err != nil {
		return bareBase64, nil // 解码失败，透传让 gopay 自己报错
	}
	// 尝试按 PKCS8 解析
	key, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		// 不是 PKCS8，原值返回（可能已是 PKCS1）
		return bareBase64, nil
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not RSA (got %T)", key)
	}
	pkcs1DER := x509.MarshalPKCS1PrivateKey(rsaKey)
	return base64.StdEncoding.EncodeToString(pkcs1DER), nil
}

// stripPEMHeaders 去掉 PEM 的 BEGIN/END 行和空行，返回裸 base64 字符串
func stripPEMHeaders(pem string) string {
	var lines []string
	for _, line := range strings.Split(pem, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-----") {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "")
}

// AlipayPublicKeyBare 返回支付宝公钥的裸 base64 字符串，供 handler 层验签使用
func (s *AlipayService) AlipayPublicKeyBare(ctx context.Context) (string, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return "", err
	}
	return stripPEMHeaders(strings.ReplaceAll(cfg.PublicKey, `\n`, "\n")), nil
}
