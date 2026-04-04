-- 085_create_wechat_pay_orders.sql
-- 创建微信支付订单表，用于记录用户扫码支付充值记录

CREATE TABLE IF NOT EXISTS wechat_pay_orders (
    id          BIGSERIAL PRIMARY KEY,
    order_no    VARCHAR(64)     NOT NULL UNIQUE,
    user_id     BIGINT          NOT NULL,
    package_id  INTEGER         NOT NULL,
    cny_fee     INTEGER         NOT NULL,        -- 人民币金额（分）
    usd_amount  DECIMAL(20, 8)  NOT NULL,        -- 到账美元
    status      VARCHAR(20)     NOT NULL DEFAULT 'pending',
    wechat_trade_no VARCHAR(64),
    code_url    TEXT,
    expires_at  TIMESTAMPTZ     NOT NULL,
    paid_at     TIMESTAMPTZ,
    notify_data TEXT,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wechat_pay_orders_user_id         ON wechat_pay_orders (user_id);
CREATE INDEX IF NOT EXISTS idx_wechat_pay_orders_status          ON wechat_pay_orders (status);
CREATE INDEX IF NOT EXISTS idx_wechat_pay_orders_user_created_at ON wechat_pay_orders (user_id, created_at);
