-- 083_backfill_allowed_groups_from_subscriptions.sql
-- 此迁移的逻辑已合并到 082_subscription_plan_refactor.sql（步骤 5）
-- 使用 source_group_id 做精确 ID 映射，比原来的 name 匹配更可靠
-- 保留此文件以避免迁移编号跳跃

-- no-op: 回填已在 082 中通过 source_group_id 完成
SELECT 1;
