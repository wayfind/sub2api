-- 082_subscription_plan_refactor.sql
-- 订阅系统重构：将订阅从分组解耦为独立实体
--
-- 变更：
-- 1. 创建 subscription_plans 表
-- 2. user_subscriptions: group_id → plan_id
-- 3. redeem_codes: group_id → plan_id
-- 4. 从现有 subscription 类型分组迁移数据
-- 5. 清理 groups 表中的订阅相关字段

-- ============================================================
-- 1. 创建 subscription_plans 表
-- ============================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id              BIGSERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    description     TEXT,
    visibility      VARCHAR(20) NOT NULL DEFAULT 'public',
    status          VARCHAR(20) NOT NULL DEFAULT 'active',
    daily_limit_usd DECIMAL(20,8),
    weekly_limit_usd DECIMAL(20,8),
    monthly_limit_usd DECIMAL(20,8),
    default_validity_days INT NOT NULL DEFAULT 30,
    price           DECIMAL(20,8),
    sort_order      INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_plans_name_active
    ON subscription_plans(name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_subscription_plans_visibility_status
    ON subscription_plans(visibility, status);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_status
    ON subscription_plans(status);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_sort_order
    ON subscription_plans(sort_order);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_deleted_at
    ON subscription_plans(deleted_at);

-- ============================================================
-- 2. 从现有 subscription 类型分组迁移到 subscription_plans
-- ============================================================

-- 添加临时列记录源 group_id
ALTER TABLE subscription_plans ADD COLUMN source_group_id BIGINT;

-- 从 subscription 类型分组迁移，保留原始 group_id 映射
INSERT INTO subscription_plans (name, description, visibility, status, daily_limit_usd, weekly_limit_usd, monthly_limit_usd, default_validity_days, sort_order, created_at, updated_at, source_group_id)
SELECT
    g.name,
    g.description,
    'private',
    g.status,
    g.daily_limit_usd,
    g.weekly_limit_usd,
    g.monthly_limit_usd,
    g.default_validity_days,
    g.sort_order,
    g.created_at,
    g.updated_at,
    g.id
FROM groups g
WHERE g.subscription_type = 'subscription'
  AND g.deleted_at IS NULL;

-- ============================================================
-- 3. user_subscriptions: group_id → plan_id
-- ============================================================

-- 添加 plan_id 列
ALTER TABLE user_subscriptions ADD COLUMN IF NOT EXISTS plan_id BIGINT;

-- 填充 plan_id：通过 source_group_id 做精确 ID 映射
UPDATE user_subscriptions us
SET plan_id = sp.id
FROM subscription_plans sp
WHERE sp.source_group_id = us.group_id
  AND us.plan_id IS NULL;

-- 对于无法匹配的记录（包括软删除的），创建一个 fallback plan
DO $$
DECLARE
    fallback_plan_id BIGINT;
BEGIN
    IF EXISTS (SELECT 1 FROM user_subscriptions WHERE plan_id IS NULL) THEN
        INSERT INTO subscription_plans (name, description, visibility, status)
        VALUES ('_migrated_fallback', 'Auto-created fallback plan during migration', 'private', 'inactive')
        RETURNING id INTO fallback_plan_id;

        UPDATE user_subscriptions SET plan_id = fallback_plan_id WHERE plan_id IS NULL;
    END IF;
END $$;

-- 设为 NOT NULL
ALTER TABLE user_subscriptions ALTER COLUMN plan_id SET NOT NULL;

-- 添加外键
ALTER TABLE user_subscriptions ADD CONSTRAINT fk_user_subscriptions_plan_id
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id) ON DELETE CASCADE;

-- 删除旧的唯一约束和索引
DROP INDEX IF EXISTS idx_user_subscriptions_group_id;
DROP INDEX IF EXISTS user_subscriptions_user_id_group_id;

-- 创建新索引
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_plan_id ON user_subscriptions(plan_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_sub_user_plan_active
    ON user_subscriptions(user_id, plan_id) WHERE deleted_at IS NULL;

-- 删除 group_id 列和外键
ALTER TABLE user_subscriptions DROP CONSTRAINT IF EXISTS user_subscriptions_groups_subscriptions;
ALTER TABLE user_subscriptions DROP COLUMN IF EXISTS group_id;

-- ============================================================
-- 4. redeem_codes: group_id → plan_id
-- ============================================================

-- 添加 plan_id 列
ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS plan_id BIGINT;

-- 填充 plan_id：通过 source_group_id 做精确 ID 映射
UPDATE redeem_codes rc
SET plan_id = sp.id
FROM subscription_plans sp
WHERE sp.source_group_id = rc.group_id
  AND rc.plan_id IS NULL;

-- 添加外键
ALTER TABLE redeem_codes ADD CONSTRAINT fk_redeem_codes_plan_id
    FOREIGN KEY (plan_id) REFERENCES subscription_plans(id) ON DELETE SET NULL;

-- 删除旧索引
DROP INDEX IF EXISTS idx_redeem_codes_group_id;
DROP INDEX IF EXISTS redeem_codes_group_id;

-- 创建新索引
CREATE INDEX IF NOT EXISTS idx_redeem_codes_plan_id ON redeem_codes(plan_id);

-- 删除 group_id 列和外键
ALTER TABLE redeem_codes DROP CONSTRAINT IF EXISTS redeem_codes_groups_redeem_codes;
ALTER TABLE redeem_codes DROP COLUMN IF EXISTS group_id;

-- ============================================================
-- 5. 回填 user_allowed_groups（趁 source_group_id 还在）
-- 订阅解耦后，用户通过订阅获得的 exclusive 分组访问权需要显式记录
-- ============================================================
INSERT INTO user_allowed_groups (user_id, group_id)
SELECT DISTINCT us.user_id, sp.source_group_id
FROM user_subscriptions us
JOIN subscription_plans sp ON sp.id = us.plan_id AND sp.deleted_at IS NULL
JOIN groups g ON g.id = sp.source_group_id AND g.deleted_at IS NULL AND g.is_exclusive = true
WHERE us.deleted_at IS NULL
ON CONFLICT (user_id, group_id) DO NOTHING;

-- ============================================================
-- 6. 清理临时列 source_group_id
-- ============================================================
ALTER TABLE subscription_plans DROP COLUMN source_group_id;

-- ============================================================
-- 7. 清理 groups 表订阅相关字段
-- ============================================================
DROP INDEX IF EXISTS idx_groups_subscription_type;
ALTER TABLE groups DROP COLUMN IF EXISTS subscription_type;
ALTER TABLE groups DROP COLUMN IF EXISTS daily_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS weekly_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS monthly_limit_usd;
ALTER TABLE groups DROP COLUMN IF EXISTS default_validity_days;
