-- 090_add_pricing_snapshot.sql
-- 在 usage_logs 表新增定价快照和有效折扣率字段，用于审计和数据订正。
-- pricing_snapshot: 当时实际使用的 per-token 单价 (U) + 来源
-- effective_rate: 当时实际生效的折扣率（替代不可靠的 rate_multiplier 历史数据）

ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS pricing_snapshot JSONB;
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS effective_rate DECIMAL(10,4);
