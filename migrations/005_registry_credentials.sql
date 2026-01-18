-- Registry Credentials table
CREATE TABLE IF NOT EXISTS registry_credentials (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id VARCHAR(100) NOT NULL,
    org_id VARCHAR(100),
    name VARCHAR(100) NOT NULL COMMENT '凭证名称',
    registry VARCHAR(255) NOT NULL DEFAULT 'docker.io' COMMENT '仓库地址',
    username VARCHAR(255) NOT NULL,
    password_encrypted VARCHAR(1000) NOT NULL COMMENT '加密后的密码/token',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_user_name (user_id, name),
    INDEX idx_user (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Add registry_credential_id to user_endpoints
ALTER TABLE user_endpoints ADD COLUMN registry_credential_id BIGINT DEFAULT NULL COMMENT '镜像仓库凭证ID';
