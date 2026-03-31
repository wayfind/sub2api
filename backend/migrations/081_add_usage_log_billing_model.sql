-- 新增 billing_model 列，记录实际用于定价查询的模型名
-- 可能与 model（客户端请求）和 upstream_model（上游返回）不同
-- 用于计费协议与定价模型解耦场景（如 Anthropic 协议代理 MiniMax 模型）

SET LOCAL lock_timeout = '5s';
SET LOCAL statement_timeout = '30s';

ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS billing_model TEXT;

COMMENT ON COLUMN usage_logs.billing_model IS '实际用于定价查询的模型名（可能与 model/upstream_model 不同）';
