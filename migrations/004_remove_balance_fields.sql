-- 移除 user_balances 表中不再需要的字段 (余额从主站获取)
ALTER TABLE user_balances 
    DROP COLUMN IF EXISTS balance,
    DROP COLUMN IF EXISTS credit_limit,
    DROP COLUMN IF EXISTS currency,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS low_balance_threshold;

-- 移除 billing_transactions 表中不再需要的字段
ALTER TABLE billing_transactions
    DROP COLUMN IF EXISTS balance_before,
    DROP COLUMN IF EXISTS balance_after,
    DROP COLUMN IF EXISTS currency;

-- 删除 recharge_records 表 (充值由主站管理)
DROP TABLE IF EXISTS recharge_records;
