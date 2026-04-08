-- 089_convert_usage_logs_to_u_tokens.sql
-- usage_logs 费用字段 ×10（CNY 1:1 → U 代币）
--
-- 幂等条件：actual_cost < 15。
-- 单次 API 请求的 CNY 费用最高不超过几元，
-- ×10 后最小变为 15+，因此 < 15 可靠区分未迁移行。

UPDATE usage_logs SET
  input_cost = input_cost * 10,
  output_cost = output_cost * 10,
  cache_creation_cost = cache_creation_cost * 10,
  cache_read_cost = cache_read_cost * 10,
  total_cost = total_cost * 10,
  actual_cost = actual_cost * 10
WHERE actual_cost != 0 AND actual_cost < 15;
