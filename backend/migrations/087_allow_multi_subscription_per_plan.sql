-- 087: 允许同一用户对同一 plan 拥有多条活跃订阅（叠加购买）
--
-- 原约束 idx_user_sub_user_plan_active 限制了 (user_id, plan_id) 唯一，
-- 新业务允许重复购买叠加，需要移除该约束。
-- 保留普通索引以维持查询性能。

DROP INDEX IF EXISTS idx_user_sub_user_plan_active;

-- 保留非唯一复合索引支持 (user_id, plan_id) 的查询
CREATE INDEX IF NOT EXISTS idx_user_sub_user_plan ON user_subscriptions(user_id, plan_id) WHERE deleted_at IS NULL;
