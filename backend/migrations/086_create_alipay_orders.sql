-- 支付宝订单表
CREATE TABLE IF NOT EXISTS alipay_orders (
    id              BIGSERIAL PRIMARY KEY,
    order_no        VARCHAR(64)    NOT NULL UNIQUE,
    user_id         BIGINT         NOT NULL,
    package_id      INTEGER        NOT NULL,
    cny_fee         INTEGER        NOT NULL,
    usd_amount      DECIMAL(20,8)  NOT NULL,
    status          VARCHAR(20)    NOT NULL DEFAULT 'pending',
    alipay_trade_no VARCHAR(64),
    qr_code         TEXT,
    expires_at      TIMESTAMPTZ    NOT NULL,
    paid_at         TIMESTAMPTZ,
    notify_data     TEXT,
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS alipay_orders_user_id_idx        ON alipay_orders (user_id);
CREATE INDEX IF NOT EXISTS alipay_orders_status_idx         ON alipay_orders (status);
CREATE INDEX IF NOT EXISTS alipay_orders_user_created_idx   ON alipay_orders (user_id, created_at);
