-- Portal 数据库迁移: 价格字段改为 BIGINT, 删除不需要的表
-- 创建时间: 2026-01-16
-- 价格单位: 1000000 = 1 USD

-- 1. 删除不需要的表
DROP TABLE IF EXISTS worker_billing_state;

-- 2. 修改 user_balances 表
ALTER TABLE user_balances
    MODIFY COLUMN balance BIGINT NOT NULL DEFAULT 0 COMMENT '余额 (1000000 = 1 USD)',
    MODIFY COLUMN credit_limit BIGINT DEFAULT 0 COMMENT '信用额度',
    MODIFY COLUMN low_balance_threshold BIGINT DEFAULT 10000000 COMMENT '低余额阈值 (默认 10 USD)';

-- 3. 修改 recharge_records 表
ALTER TABLE recharge_records
    MODIFY COLUMN amount BIGINT NOT NULL COMMENT '充值金额 (1000000 = 1 USD)';

-- 4. 修改 spec_pricing 表
ALTER TABLE spec_pricing
    CHANGE COLUMN default_price_per_hour price_per_hour BIGINT NOT NULL COMMENT '每小时价格 (1000000 = 1 USD)',
    MODIFY COLUMN min_price BIGINT COMMENT '最低价格',
    MODIFY COLUMN max_price BIGINT COMMENT '最高价格';

-- 5. 修改 user_endpoints 表
ALTER TABLE user_endpoints
    MODIFY COLUMN price_per_hour BIGINT COMMENT '每小时价格 (1000000 = 1 USD)';

-- 6. 修改 cluster_pricing_overrides 表
ALTER TABLE cluster_pricing_overrides
    MODIFY COLUMN price_per_hour BIGINT NOT NULL COMMENT '每小时价格 (1000000 = 1 USD)';

-- 7. 重建 billing_transactions 表
DROP TABLE IF EXISTS billing_transactions;
CREATE TABLE billing_transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    endpoint_id BIGINT NOT NULL,
    cluster_id VARCHAR(100) NOT NULL,
    worker_id VARCHAR(255) NOT NULL,
    gpu_type VARCHAR(100),
    gpu_count INT DEFAULT 0,
    billing_period_start TIMESTAMP NOT NULL,
    billing_period_end TIMESTAMP NOT NULL,
    duration_seconds BIGINT NOT NULL COMMENT '计费时长(秒)',
    price_per_hour BIGINT NOT NULL COMMENT '每小时价格 (1000000 = 1 USD)',
    amount BIGINT NOT NULL COMMENT '扣费金额 (1000000 = 1 USD)',
    balance_before BIGINT COMMENT '扣费前余额',
    balance_after BIGINT COMMENT '扣费后余额',
    currency VARCHAR(10) DEFAULT 'USD',
    status VARCHAR(50) DEFAULT 'success' COMMENT 'success, insufficient_balance',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_billing (user_id, created_at),
    INDEX idx_endpoint_billing (endpoint_id, created_at),
    INDEX idx_worker_billing (worker_id, billing_period_start)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计费流水表';

-- 8. 修改 user_preferences 表
ALTER TABLE user_preferences
    MODIFY COLUMN daily_budget_limit BIGINT COMMENT '每日预算上限 (1000000 = 1 USD)',
    MODIFY COLUMN monthly_budget_limit BIGINT COMMENT '每月预算上限 (1000000 = 1 USD)';

-- 9. 更新 spec_pricing 初始数据 (价格 * 1000000)
UPDATE spec_pricing SET 
    price_per_hour = 2800000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'GPU-A100-40GB';

UPDATE spec_pricing SET 
    price_per_hour = 4500000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'GPU-H100-80GB';

UPDATE spec_pricing SET 
    price_per_hour = 1200000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'GPU-A10-24GB';

UPDATE spec_pricing SET 
    price_per_hour = 100000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'CPU-4C-8G';

UPDATE spec_pricing SET 
    price_per_hour = 200000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'CPU-8C-16G';

UPDATE spec_pricing SET 
    price_per_hour = 400000,
    min_price = NULL,
    max_price = NULL
WHERE spec_name = 'CPU-16C-32G';
