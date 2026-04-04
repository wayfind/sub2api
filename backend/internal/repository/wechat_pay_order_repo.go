package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/wechatpayorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type wechatPayOrderRepository struct {
	client *dbent.Client
}

func NewWechatPayOrderRepository(client *dbent.Client) service.WechatPayOrderRepository {
	return &wechatPayOrderRepository{client: client}
}

func (r *wechatPayOrderRepository) Create(ctx context.Context, order *service.WechatPayOrder) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.WechatPayOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetPackageID(order.PackageID).
		SetCnyFee(order.CnyFee).
		SetUsdAmount(order.UsdAmount).
		SetStatus(order.Status).
		SetNillableCodeURL(order.CodeURL).
		SetExpiresAt(order.ExpiresAt).
		Save(ctx)
	if err != nil {
		return err
	}
	order.ID = created.ID
	order.CreatedAt = created.CreatedAt
	order.UpdatedAt = created.UpdatedAt
	return nil
}

func (r *wechatPayOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.WechatPayOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.WechatPayOrder.Query().
		Where(wechatpayorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrWechatPayOrderNotFound
		}
		return nil, err
	}
	return wechatPayOrderEntityToService(m), nil
}

func (r *wechatPayOrderRepository) GetByID(ctx context.Context, id int64) (*service.WechatPayOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.WechatPayOrder.Query().
		Where(wechatpayorder.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrWechatPayOrderNotFound
		}
		return nil, err
	}
	return wechatPayOrderEntityToService(m), nil
}

// MarkPaid 将订单标记为已支付，同一订单只能成功一次（幂等：仅更新 pending 状态的记录）。
// 不检查 expires_at：微信回调可能在本地订单"过期"后才到达，只要钱已付就应入账。
// 支持事务上下文（通过 ent.NewTxContext 注入），调用方可在事务中同时完成余额更新。
func (r *wechatPayOrderRepository) MarkPaid(ctx context.Context, orderNo, wechatTradeNo, notifyData string) (bool, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.WechatPayOrder.Update().
		Where(
			wechatpayorder.OrderNoEQ(orderNo),
			wechatpayorder.StatusEQ("pending"),
		).
		SetStatus("paid").
		SetWechatTradeNo(wechatTradeNo).
		SetPaidAt(time.Now()).
		SetNotifyData(notifyData).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *wechatPayOrderRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.WechatPayOrder, *pagination.PaginationResult, error) {
	q := r.client.WechatPayOrder.Query().
		Where(wechatpayorder.UserIDEQ(userID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(wechatpayorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return wechatPayOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *wechatPayOrderRepository) List(ctx context.Context, params pagination.PaginationParams, status string) ([]service.WechatPayOrder, *pagination.PaginationResult, error) {
	q := r.client.WechatPayOrder.Query()
	if status != "" {
		q = q.Where(wechatpayorder.StatusEQ(status))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(wechatpayorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return wechatPayOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func wechatPayOrderEntityToService(m *dbent.WechatPayOrder) *service.WechatPayOrder {
	if m == nil {
		return nil
	}
	return &service.WechatPayOrder{
		ID:             m.ID,
		OrderNo:        m.OrderNo,
		UserID:         m.UserID,
		PackageID:      m.PackageID,
		CnyFee:         m.CnyFee,
		UsdAmount:      m.UsdAmount,
		Status:         m.Status,
		WechatTradeNo:  m.WechatTradeNo,
		CodeURL:        m.CodeURL,
		ExpiresAt:      m.ExpiresAt,
		PaidAt:         m.PaidAt,
		NotifyData:     m.NotifyData,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func wechatPayOrderEntitiesToService(models []*dbent.WechatPayOrder) []service.WechatPayOrder {
	out := make([]service.WechatPayOrder, 0, len(models))
	for _, m := range models {
		if s := wechatPayOrderEntityToService(m); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
