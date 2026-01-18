-- Portal 数据库迁移: 添加 Workers 和监控表
-- 创建时间: 2026-01-16

-- 1. 更新 task_routing 表，添加 input 和 worker_id 字段
ALTER TABLE task_routing 
    ADD COLUMN input JSON COMMENT '任务输入',
    ADD COLUMN worker_id VARCHAR(255) COMMENT '执行的 Worker ID',
    ADD COLUMN status VARCHAR(50) DEFAULT 'PENDING' COMMENT '任务状态',
    ADD COLUMN created_at TIMESTAMP NULL COMMENT '任务创建时间',
    ADD COLUMN completed_at TIMESTAMP NULL COMMENT '任务完成时间',
    ADD COLUMN execution_time_ms BIGINT DEFAULT 0 COMMENT '执行时长(ms)',
    ADD INDEX idx_status (status),
    ADD INDEX idx_endpoint (endpoint_id, status);

-- 2. Workers 表 - 本地缓存 Worker 信息
CREATE TABLE workers (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    worker_id VARCHAR(255) NOT NULL UNIQUE COMMENT 'Worker ID (Pod name)',
    endpoint_id BIGINT NOT NULL COMMENT 'Portal Endpoint ID',
    cluster_id VARCHAR(100) NOT NULL COMMENT '所属集群',
    user_id VARCHAR(100) NOT NULL COMMENT '用户 ID',
    
    -- Worker 状态
    pod_name VARCHAR(255) COMMENT 'Pod 名称',
    status VARCHAR(50) NOT NULL DEFAULT 'STARTING' COMMENT '状态: STARTING, ONLINE, BUSY, DRAINING, OFFLINE',
    
    -- 生命周期
    pod_created_at TIMESTAMP NULL COMMENT 'Pod 创建时间',
    pod_started_at TIMESTAMP NULL COMMENT 'Pod 启动时间(计费起点)',
    pod_ready_at TIMESTAMP NULL COMMENT 'Pod Ready 时间',
    pod_terminated_at TIMESTAMP NULL COMMENT 'Pod 终止时间',
    cold_start_duration_ms BIGINT COMMENT '冷启动时长(ms)',
    
    -- 任务统计
    current_jobs INT DEFAULT 0 COMMENT '当前任务数',
    total_tasks_completed BIGINT DEFAULT 0 COMMENT '完成任务数',
    total_tasks_failed BIGINT DEFAULT 0 COMMENT '失败任务数',
    total_execution_time_ms BIGINT DEFAULT 0 COMMENT '总执行时长(ms)',
    last_task_time TIMESTAMP NULL COMMENT '最后任务时间',
    last_heartbeat TIMESTAMP NULL COMMENT '最后心跳时间',
    
    -- 计费相关 (价格单位: 1000000 = 1 USD)
    billing_status VARCHAR(50) DEFAULT 'pending' COMMENT '计费状态: pending, active, final_billed',
    last_billed_at TIMESTAMP NULL COMMENT '上次计费时间',
    total_billed_seconds BIGINT DEFAULT 0 COMMENT '已计费时长(秒)',
    total_billed_amount BIGINT DEFAULT 0 COMMENT '已计费金额 (1000000 = 1 USD)',
    
    -- 同步信息
    last_synced_at TIMESTAMP NULL COMMENT '最后同步时间',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_endpoint (endpoint_id),
    INDEX idx_cluster (cluster_id),
    INDEX idx_user (user_id),
    INDEX idx_status (status),
    INDEX idx_billing_status (billing_status),
    FOREIGN KEY (endpoint_id) REFERENCES user_endpoints(id) ON DELETE CASCADE,
    FOREIGN KEY (cluster_id) REFERENCES clusters(cluster_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Worker 信息表';

-- 3. Endpoint 分钟级统计表
CREATE TABLE endpoint_minute_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    endpoint_id BIGINT NOT NULL COMMENT 'Portal Endpoint ID',
    stat_minute DATETIME NOT NULL COMMENT '统计分钟',
    
    -- Worker 统计
    active_workers INT DEFAULT 0,
    idle_workers INT DEFAULT 0,
    
    -- 任务统计
    tasks_submitted INT DEFAULT 0,
    tasks_completed INT DEFAULT 0,
    tasks_failed INT DEFAULT 0,
    tasks_timeout INT DEFAULT 0,
    
    -- 性能指标
    avg_queue_wait_ms DECIMAL(10,2) DEFAULT 0,
    avg_execution_ms DECIMAL(10,2) DEFAULT 0,
    p95_execution_ms DECIMAL(10,2) DEFAULT 0,
    
    -- Worker 生命周期
    workers_created INT DEFAULT 0,
    workers_terminated INT DEFAULT 0,
    cold_starts INT DEFAULT 0,
    avg_cold_start_ms DECIMAL(10,2) DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_endpoint_minute (endpoint_id, stat_minute),
    INDEX idx_stat_minute (stat_minute),
    FOREIGN KEY (endpoint_id) REFERENCES user_endpoints(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Endpoint 分钟级统计';

-- 4. Endpoint 小时级统计表
CREATE TABLE endpoint_hourly_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    endpoint_id BIGINT NOT NULL COMMENT 'Portal Endpoint ID',
    stat_hour DATETIME NOT NULL COMMENT '统计小时',
    
    -- Worker 统计
    active_workers INT DEFAULT 0,
    idle_workers INT DEFAULT 0,
    avg_workers DECIMAL(10,2) DEFAULT 0,
    max_workers INT DEFAULT 0,
    
    -- 任务统计
    tasks_submitted INT DEFAULT 0,
    tasks_completed INT DEFAULT 0,
    tasks_failed INT DEFAULT 0,
    tasks_timeout INT DEFAULT 0,
    
    -- 性能指标
    avg_queue_wait_ms DECIMAL(10,2) DEFAULT 0,
    avg_execution_ms DECIMAL(10,2) DEFAULT 0,
    p50_execution_ms DECIMAL(10,2) DEFAULT 0,
    p95_execution_ms DECIMAL(10,2) DEFAULT 0,
    
    -- Worker 生命周期
    workers_created INT DEFAULT 0,
    workers_terminated INT DEFAULT 0,
    cold_starts INT DEFAULT 0,
    avg_cold_start_ms DECIMAL(10,2) DEFAULT 0,
    
    -- 费用统计 (价格单位: 1000000 = 1 USD)
    total_worker_seconds BIGINT DEFAULT 0 COMMENT '总 Worker 运行秒数',
    total_cost BIGINT DEFAULT 0 COMMENT '总费用 (1000000 = 1 USD)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_endpoint_hour (endpoint_id, stat_hour),
    INDEX idx_stat_hour (stat_hour),
    FOREIGN KEY (endpoint_id) REFERENCES user_endpoints(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Endpoint 小时级统计';

-- 5. Endpoint 日级统计表
CREATE TABLE endpoint_daily_stats (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    endpoint_id BIGINT NOT NULL COMMENT 'Portal Endpoint ID',
    stat_date DATE NOT NULL COMMENT '统计日期',
    
    -- Worker 统计
    avg_workers DECIMAL(10,2) DEFAULT 0,
    max_workers INT DEFAULT 0,
    peak_hour TINYINT COMMENT '峰值小时',
    
    -- 任务统计
    tasks_submitted INT DEFAULT 0,
    tasks_completed INT DEFAULT 0,
    tasks_failed INT DEFAULT 0,
    tasks_timeout INT DEFAULT 0,
    success_rate DECIMAL(5,2) DEFAULT 0 COMMENT '成功率(%)',
    
    -- 性能指标
    avg_queue_wait_ms DECIMAL(10,2) DEFAULT 0,
    avg_execution_ms DECIMAL(10,2) DEFAULT 0,
    p50_execution_ms DECIMAL(10,2) DEFAULT 0,
    p95_execution_ms DECIMAL(10,2) DEFAULT 0,
    
    -- Worker 生命周期
    workers_created INT DEFAULT 0,
    workers_terminated INT DEFAULT 0,
    cold_starts INT DEFAULT 0,
    avg_cold_start_ms DECIMAL(10,2) DEFAULT 0,
    
    -- 费用统计 (价格单位: 1000000 = 1 USD)
    total_worker_hours DECIMAL(10,2) DEFAULT 0 COMMENT '总 Worker 运行小时',
    total_cost BIGINT DEFAULT 0 COMMENT '总费用 (1000000 = 1 USD)',
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY uk_endpoint_date (endpoint_id, stat_date),
    INDEX idx_stat_date (stat_date),
    FOREIGN KEY (endpoint_id) REFERENCES user_endpoints(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Endpoint 日级统计';
