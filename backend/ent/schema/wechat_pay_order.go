package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WechatPayOrder holds the schema definition for the WechatPayOrder entity.
//
// 删除策略：硬删除
// 订单记录通过 status 字段追踪生命周期，已完成/过期的订单无需软删除。
type WechatPayOrder struct {
	ent.Schema
}

func (WechatPayOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "wechat_pay_orders"},
	}
}

func (WechatPayOrder) Fields() []ent.Field {
	return []ent.Field{
		// 业务订单号，格式: WX{timestamp}{random6}，用于微信支付 out_trade_no
		field.String("order_no").
			MaxLen(64).
			NotEmpty().
			Unique(),
		// 下单用户 ID
		field.Int64("user_id"),
		// 套餐 ID（来自 Setting 表 wechat_pay_packages 中的 id）
		field.Int("package_id"),
		// 支付金额（人民币，单位：分）
		field.Int("cny_fee"),
		// 到账美元金额
		field.Float("usd_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		// 订单状态：pending / paid / expired / refunded
		field.String("status").
			MaxLen(20).
			Default("pending"),
		// 微信支付流水号（支付成功后填入）
		field.String("wechat_trade_no").
			MaxLen(64).
			Optional().
			Nillable(),
		// 微信返回的二维码链接
		field.String("code_url").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		// 订单过期时间（30分钟）
		field.Time("expires_at").
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 支付成功时间
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		// 微信回调原始数据（审计用）
		field.String("notify_data").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (WechatPayOrder) Edges() []ent.Edge {
	return nil
}

func (WechatPayOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("order_no"),
		index.Fields("status"),
		index.Fields("user_id", "created_at"),
	}
}
