-- 083_backfill_allowed_groups_from_subscriptions.sql
-- 修复：订阅解耦后，用户丢失了通过订阅获得的分组访问权
-- 根据用户的活跃订阅，将对应的分组添加到 user_allowed_groups

-- 通过 plan 名与 group 名匹配，为每个有活跃订阅的用户补充分组权限
INSERT INTO user_allowed_groups (user_id, group_id)
SELECT DISTINCT us.user_id, g.id
FROM user_subscriptions us
JOIN subscription_plans sp ON sp.id = us.plan_id AND sp.deleted_at IS NULL
JOIN groups g ON g.name = sp.name AND g.deleted_at IS NULL AND g.is_exclusive = true
WHERE us.status = 'active'
  AND us.deleted_at IS NULL
  AND us.expires_at > NOW()
ON CONFLICT (user_id, group_id) DO NOTHING;
