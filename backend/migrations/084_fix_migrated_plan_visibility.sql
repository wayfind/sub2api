-- 084_fix_migrated_plan_visibility.sql
-- 修复：082 迁移时将所有订阅计划可见性硬编码为 'public'
-- 从旧的 exclusive 分组迁移来的计划应为 'private'（管理员分配）

UPDATE subscription_plans
SET visibility = 'private', updated_at = NOW()
WHERE visibility = 'public'
  AND deleted_at IS NULL
  AND name != '_migrated_fallback';
