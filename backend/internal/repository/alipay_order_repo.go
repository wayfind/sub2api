package repository

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/alipayorder"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type alipayOrderRepository struct {
	client *dbent.Client
}

func NewAlipayOrderRepository(client *dbent.Client) service.AlipayOrderRepository {
	return &alipayOrderRepository{client: client}
}

func (r *alipayOrderRepository) Create(ctx context.Context, order *service.AlipayOrder) error {
	client := clientFromContext(ctx, r.client)
	created, err := client.AlipayOrder.Create().
		SetOrderNo(order.OrderNo).
		SetUserID(order.UserID).
		SetPackageID(order.PackageID).
		SetCnyFee(order.CnyFee).
		SetUsdAmount(order.UsdAmount).
		SetStatus(order.Status).
		SetNillableQrCode(order.QRCode).
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

func (r *alipayOrderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*service.AlipayOrder, error) {
	client := clientFromContext(ctx, r.client)
	m, err := client.AlipayOrder.Query().
		Where(alipayorder.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrAlipayOrderNotFound
		}
		return nil, err
	}
	return alipayOrderEntityToService(m), nil
}

// MarkPaid 幂等标记支付成功，仅更新 pending 状态的记录。
// 不检查 expires_at：支付宝回调可能在本地订单"过期"后才到达。
func (r *alipayOrderRepository) MarkPaid(ctx context.Context, orderNo, alipayTradeNo, notifyData string) (bool, error) {
	client := clientFromContext(ctx, r.client)
	n, err := client.AlipayOrder.Update().
		Where(
			alipayorder.OrderNoEQ(orderNo),
			alipayorder.StatusEQ("pending"),
		).
		SetStatus("paid").
		SetAlipayTradeNo(alipayTradeNo).
		SetPaidAt(time.Now()).
		SetNotifyData(notifyData).
		Save(ctx)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *alipayOrderRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.AlipayOrder, *pagination.PaginationResult, error) {
	q := r.client.AlipayOrder.Query().
		Where(alipayorder.UserIDEQ(userID))

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(alipayorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return alipayOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func (r *alipayOrderRepository) List(ctx context.Context, params pagination.PaginationParams, status string) ([]service.AlipayOrder, *pagination.PaginationResult, error) {
	q := r.client.AlipayOrder.Query()
	if status != "" {
		q = q.Where(alipayorder.StatusEQ(status))
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}

	orders, err := q.
		Offset(params.Offset()).
		Limit(params.Limit()).
		Order(dbent.Desc(alipayorder.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, nil, err
	}

	return alipayOrderEntitiesToService(orders), paginationResultFromTotal(int64(total), params), nil
}

func alipayOrderEntityToService(m *dbent.AlipayOrder) *service.AlipayOrder {
	if m == nil {
		return nil
	}
	return &service.AlipayOrder{
		ID:            m.ID,
		OrderNo:       m.OrderNo,
		UserID:        m.UserID,
		PackageID:     m.PackageID,
		CnyFee:        m.CnyFee,
		UsdAmount:     m.UsdAmount,
		Status:        m.Status,
		AlipayTradeNo: m.AlipayTradeNo,
		QRCode:        m.QrCode,
		ExpiresAt:     m.ExpiresAt,
		PaidAt:        m.PaidAt,
		NotifyData:    m.NotifyData,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func alipayOrderEntitiesToService(models []*dbent.AlipayOrder) []service.AlipayOrder {
	out := make([]service.AlipayOrder, 0, len(models))
	for _, m := range models {
		if s := alipayOrderEntityToService(m); s != nil {
			out = append(out, *s)
		}
	}
	return out
}
