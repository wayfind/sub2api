-- 088_convert_usd_to_u_tokens.sql
-- 将系统所有金额从 CNY（1:1 存储）转为 U 代币单位（1 CNY = 10 U）
-- 系统历史上所有金额字段虽然叫 "usd"，实际存储的是 1:1 CNY 值。
-- 此迁移不可逆（无 down）。执行前务必备份数据库。
-- 注意：migration runner 会自动包事务，不要在文件中写 BEGIN/COMMIT。

-- 1. 用户余额 ×10
UPDATE users SET balance = balance * 10 WHERE balance != 0 AND deleted_at IS NULL;

-- 2. usage_logs 在 089 中单独处理（大表独立迁移）

-- 3. 订阅计划限额 ×10
UPDATE subscription_plans SET
  daily_limit_usd = daily_limit_usd * 10,
  weekly_limit_usd = weekly_limit_usd * 10,
  monthly_limit_usd = monthly_limit_usd * 10
WHERE deleted_at IS NULL
  AND (daily_limit_usd IS NOT NULL OR weekly_limit_usd IS NOT NULL OR monthly_limit_usd IS NOT NULL);

-- 3.1 订阅计划价格 ×10
UPDATE subscription_plans SET price = price * 10
WHERE deleted_at IS NULL AND price IS NOT NULL AND price != 0;

-- 4. 用户订阅已用额度 ×10
UPDATE user_subscriptions SET
  daily_usage_usd = daily_usage_usd * 10,
  weekly_usage_usd = weekly_usage_usd * 10,
  monthly_usage_usd = monthly_usage_usd * 10
WHERE deleted_at IS NULL
  AND (daily_usage_usd != 0 OR weekly_usage_usd != 0 OR monthly_usage_usd != 0);

-- 5. API Key 配额 ×10
UPDATE api_keys SET
  quota = quota * 10,
  quota_used = quota_used * 10
WHERE deleted_at IS NULL
  AND (quota != 0 OR quota_used != 0);

-- 6. 支付宝订单到账金额 ×10
UPDATE alipay_orders SET usd_amount = usd_amount * 10 WHERE usd_amount != 0;

-- 7. 微信支付订单到账金额 ×10（如果表存在）
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'wechat_pay_orders') THEN
    EXECUTE 'UPDATE wechat_pay_orders SET usd_amount = usd_amount * 10 WHERE usd_amount != 0';
  END IF;
END $$;

-- 8. Account extra JSONB 中的金额字段 ×10
DO $$
DECLARE
  keys text[] := ARRAY[
    'quota_limit', 'quota_used',
    'quota_daily_limit', 'quota_daily_used',
    'quota_weekly_limit', 'quota_weekly_used',
    'window_cost_limit', 'window_cost_sticky_reserve'
  ];
  k text;
BEGIN
  FOREACH k IN ARRAY keys
  LOOP
    EXECUTE format(
      'UPDATE accounts SET extra = jsonb_set(extra, ''{%s}'', to_jsonb((extra->>''%s'')::numeric * 10))
       WHERE deleted_at IS NULL AND extra ? ''%s'' AND (extra->>''%s'')::numeric != 0',
      k, k, k, k
    );
  END LOOP;
END $$;

-- 9. 充值套餐配置 ×10
DO $$
DECLARE
  setting_key text;
  old_val text;
  new_val text;
BEGIN
  FOREACH setting_key IN ARRAY ARRAY['wechat_pay_packages']
  LOOP
    SELECT value INTO old_val FROM settings WHERE key = setting_key;
    IF old_val IS NOT NULL AND old_val != '' AND old_val != '[]' THEN
      SELECT json_agg(
        CASE
          WHEN (elem->>'usd_amount') IS NOT NULL
          THEN (elem::jsonb || jsonb_build_object('usd_amount', (elem->>'usd_amount')::numeric * 10))::json
          ELSE elem
        END
      )::text INTO new_val
      FROM json_array_elements(old_val::json) AS elem;

      IF new_val IS NOT NULL THEN
        UPDATE settings SET value = new_val WHERE key = setting_key;
      END IF;
    END IF;
  END LOOP;
END $$;

-- 10. 仪表盘聚合数据 ×10
UPDATE usage_dashboard_daily SET
  total_cost = total_cost * 10,
  actual_cost = actual_cost * 10
WHERE total_cost != 0 OR actual_cost != 0;

UPDATE usage_dashboard_hourly SET
  total_cost = total_cost * 10,
  actual_cost = actual_cost * 10
WHERE total_cost != 0 OR actual_cost != 0;

-- 11. 兑换码面值 ×10
UPDATE redeem_codes SET value = value * 10 WHERE value != 0;

-- 12. 推广码奖励 ×10（当前无数据，预防性迁移）
UPDATE promo_codes SET bonus_amount = bonus_amount * 10 WHERE bonus_amount != 0;
UPDATE promo_code_usages SET bonus_amount = bonus_amount * 10 WHERE bonus_amount != 0;

-- 13. API Key 费率限额 ×10（当前无数据，预防性迁移）
UPDATE api_keys SET
  rate_limit_5h = rate_limit_5h * 10,
  rate_limit_1d = rate_limit_1d * 10,
  rate_limit_7d = rate_limit_7d * 10,
  usage_5h = usage_5h * 10,
  usage_1d = usage_1d * 10,
  usage_7d = usage_7d * 10
WHERE deleted_at IS NULL
  AND (rate_limit_5h != 0 OR rate_limit_1d != 0 OR rate_limit_7d != 0
    OR usage_5h != 0 OR usage_1d != 0 OR usage_7d != 0);

-- 14. 分组图片/Sora 价格 ×10（当前无数据，预防性迁移）
UPDATE groups SET
  image_price_1k = image_price_1k * 10,
  image_price_2k = image_price_2k * 10,
  image_price_4k = image_price_4k * 10,
  sora_image_price_360 = sora_image_price_360 * 10,
  sora_image_price_540 = sora_image_price_540 * 10,
  sora_video_price_per_request = sora_video_price_per_request * 10,
  sora_video_price_per_request_hd = sora_video_price_per_request_hd * 10
WHERE deleted_at IS NULL
  AND (image_price_1k IS NOT NULL OR sora_image_price_360 IS NOT NULL
    OR sora_video_price_per_request IS NOT NULL);

-- 15. 分组限额/价格 ×10（已迁移到 subscription_plans，当前无数据，预防性迁移）
UPDATE groups SET
  daily_limit_usd = daily_limit_usd * 10,
  weekly_limit_usd = weekly_limit_usd * 10,
  monthly_limit_usd = monthly_limit_usd * 10,
  price = price * 10
WHERE deleted_at IS NULL
  AND (daily_limit_usd IS NOT NULL OR price != 0);
