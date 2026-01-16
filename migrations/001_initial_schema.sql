-- Portal 数据库初始化脚本
-- 创建时间: 2026-01-14

-- 用户余额表
CREATE TABLE user_balances (
    user_id VARCHAR(100) PRIMARY KEY COMMENT '主站用户 UUID',
    balance DECIMAL(12, 4) NOT NULL DEFAULT 0 COMMENT '余额',
    credit_limit DECIMAL(12, 4) DEFAULT 0 COMMENT '信用额度',
    currency VARCHAR(10) DEFAULT 'USD' COMMENT '货币',
    status VARCHAR(50) DEFAULT 'active' COMMENT '状态: active, suspended, debt',
    low_balance_threshold DECIMAL(12, 4) DEFAULT 10.00 COMMENT '低余额阈值',
    org_id VARCHAR(100) COMMENT '组织 ID',
    user_name VARCHAR(255) COMMENT '用户名',
    email VARCHAR(255) COMMENT '邮箱',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_org (org_id),
    INDEX idx_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户余额表';

-- 充值记录表
CREATE TABLE recharge_records (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(100) NOT NULL COMMENT '主站用户 UUID',
    amount DECIMAL(12, 4) NOT NULL COMMENT '充值金额',
    currency VARCHAR(10) DEFAULT 'USD' COMMENT '货币',
    payment_method VARCHAR(50) COMMENT '支付方式',
    transaction_id VARCHAR(255) COMMENT '交易 ID',
    status VARCHAR(50) DEFAULT 'pending' COMMENT '状态',
    note TEXT COMMENT '备注',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    INDEX idx_user_recharge (user_id, created_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='充值记录表';

-- 集群注册表
CREATE TABLE clusters (
    cluster_id VARCHAR(100) PRIMARY KEY,
    cluster_name VARCHAR(255) NOT NULL,
    region VARCHAR(100) NOT NULL COMMENT '区域: us-east-1',
    api_endpoint VARCHAR(500) NOT NULL COMMENT 'Waverless API 地址',
    api_key VARCHAR(255) COMMENT 'API Key',
    status VARCHAR(50) DEFAULT 'active' COMMENT '状态',
    priority INT DEFAULT 100 COMMENT '调度优先级',
    total_gpu_slots INT DEFAULT 0,
    available_gpu_slots INT DEFAULT 0,
    last_heartbeat_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_region (region),
    INDEX idx_status (status),
    INDEX idx_heartbeat (last_heartbeat_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群注册表';

-- 集群规格表 - 记录集群支持哪些规格及容量
CREATE TABLE cluster_specs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cluster_id VARCHAR(100) NOT NULL,
    cluster_spec_name VARCHAR(100) NOT NULL COMMENT '集群内部的 spec 名称',
    spec_name VARCHAR(100) NOT NULL COMMENT '关联 spec_pricing 的名称',
    total_capacity INT DEFAULT 0,
    available_capacity INT DEFAULT 0,
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_cluster_spec (cluster_id, cluster_spec_name),
    FOREIGN KEY (cluster_id) REFERENCES clusters(cluster_id) ON DELETE CASCADE,
    INDEX idx_spec_name (spec_name),
    INDEX idx_available (is_available, available_capacity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群规格表';

-- 规格价格配置表
CREATE TABLE spec_pricing (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    spec_name VARCHAR(100) NOT NULL UNIQUE,
    spec_type VARCHAR(20) NOT NULL,
    gpu_type VARCHAR(100),
    gpu_count INT DEFAULT 0,
    cpu_cores INT NOT NULL,
    ram_gb INT NOT NULL,
    disk_gb INT,
    default_price_per_hour DECIMAL(10, 4) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    min_price DECIMAL(10, 4),
    max_price DECIMAL(10, 4),
    description TEXT,
    is_available BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_spec_type (spec_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='规格价格配置表';

-- 用户 Endpoint 表
CREATE TABLE user_endpoints (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    logical_name VARCHAR(255) NOT NULL,
    physical_name VARCHAR(255) NOT NULL,
    spec_name VARCHAR(100) NOT NULL,
    spec_type VARCHAR(20) NOT NULL,
    gpu_type VARCHAR(100),
    gpu_count INT DEFAULT 0,
    cpu_cores INT NOT NULL,
    ram_gb INT NOT NULL,
    cluster_id VARCHAR(100) NOT NULL,
    replicas INT DEFAULT 0,
    min_replicas INT NOT NULL,
    max_replicas INT NOT NULL,
    current_replicas INT DEFAULT 0,
    image VARCHAR(500) NOT NULL,
    task_timeout INT DEFAULT 3600,
    env JSON,
    price_per_hour DECIMAL(10, 4),
    currency VARCHAR(10) DEFAULT 'USD',
    prefer_region VARCHAR(100),
    status VARCHAR(50) DEFAULT 'deploying',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE KEY uk_user_endpoint (user_id, logical_name),
    FOREIGN KEY (cluster_id) REFERENCES clusters(cluster_id),
    INDEX idx_user (user_id),
    INDEX idx_cluster (cluster_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户Endpoint表';

-- 计费流水表
CREATE TABLE billing_transactions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    endpoint_id BIGINT NOT NULL,
    cluster_id VARCHAR(100) NOT NULL,
    worker_id VARCHAR(255) NOT NULL,
    gpu_type VARCHAR(100) NOT NULL,
    gpu_count INT NOT NULL,
    billing_period_start TIMESTAMP NOT NULL,
    billing_period_end TIMESTAMP NOT NULL,
    duration_seconds INT NOT NULL,
    price_per_gpu_hour DECIMAL(10, 4) NOT NULL,
    gpu_hours DECIMAL(10, 4) NOT NULL,
    amount DECIMAL(12, 4) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    balance_before DECIMAL(12, 4),
    balance_after DECIMAL(12, 4),
    status VARCHAR(50) DEFAULT 'success',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_billing (user_id, created_at),
    INDEX idx_endpoint_billing (endpoint_id, created_at),
    INDEX idx_worker_billing (worker_id, billing_period_start)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计费流水表';

-- Worker 计费状态表
CREATE TABLE worker_billing_state (
    worker_id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    endpoint_id BIGINT NOT NULL,
    cluster_id VARCHAR(100) NOT NULL,
    gpu_type VARCHAR(100) NOT NULL,
    gpu_count INT NOT NULL,
    price_per_gpu_hour DECIMAL(10, 4) NOT NULL,
    pod_started_at TIMESTAMP NOT NULL,
    pod_terminated_at TIMESTAMP NULL,
    last_billed_at TIMESTAMP NOT NULL,
    total_billed_seconds INT DEFAULT 0,
    total_billed_amount DECIMAL(12, 4) DEFAULT 0,
    billing_status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_worker (user_id),
    INDEX idx_endpoint (endpoint_id),
    INDEX idx_billing_status (billing_status, last_billed_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Worker计费状态表';

-- 任务路由记录表
CREATE TABLE task_routing (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    task_id VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    endpoint_id BIGINT NOT NULL,
    cluster_id VARCHAR(100) NOT NULL,
    submitted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_task_id (task_id),
    INDEX idx_user_task (user_id, submitted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务路由记录表';

-- 初始化规格价格数据
INSERT INTO spec_pricing (spec_name, spec_type, gpu_type, gpu_count, cpu_cores, ram_gb, disk_gb, default_price_per_hour, description) VALUES
('GPU-A100-40GB', 'GPU', 'A100-40GB', 1, 16, 64, 200, 2.80, 'NVIDIA A100 40GB GPU'),
('GPU-H100-80GB', 'GPU', 'H100-80GB', 1, 32, 128, 500, 4.50, 'NVIDIA H100 80GB GPU'),
('GPU-A10-24GB', 'GPU', 'A10-24GB', 1, 8, 32, 100, 1.20, 'NVIDIA A10 24GB GPU'),
('CPU-4C-8G', 'CPU', NULL, 0, 4, 8, 50, 0.10, '4 vCPU, 8GB RAM'),
('CPU-8C-16G', 'CPU', NULL, 0, 8, 16, 100, 0.20, '8 vCPU, 16GB RAM'),
('CPU-16C-32G', 'CPU', NULL, 0, 16, 32, 200, 0.40, '16 vCPU, 32GB RAM');


-- 用户偏好设置表
CREATE TABLE user_preferences (
    user_id VARCHAR(100) PRIMARY KEY COMMENT '主站用户 UUID',
    daily_budget_limit DECIMAL(12, 4) COMMENT '每日预算上限',
    monthly_budget_limit DECIMAL(12, 4) COMMENT '每月预算上限',
    auto_suspend_on_low_balance BOOLEAN DEFAULT true COMMENT '余额不足时自动暂停',
    auto_migrate_for_price BOOLEAN DEFAULT false COMMENT '自动迁移到更便宜的集群',
    email_notifications BOOLEAN DEFAULT true COMMENT '邮件通知',
    low_balance_alert BOOLEAN DEFAULT true COMMENT '低余额告警',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户偏好设置表';

-- 集群特殊定价表
CREATE TABLE cluster_pricing_overrides (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    cluster_id VARCHAR(100) NOT NULL,
    spec_name VARCHAR(100) NOT NULL,
    price_per_hour DECIMAL(10, 4) NOT NULL,
    currency VARCHAR(10) DEFAULT 'USD',
    effective_from TIMESTAMP NULL,
    effective_until TIMESTAMP NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_cluster_spec (cluster_id, spec_name),
    FOREIGN KEY (cluster_id) REFERENCES clusters(cluster_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='集群特殊定价表';
