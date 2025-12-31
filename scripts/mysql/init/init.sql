-- Cyber Inspector 数据库初始化脚本
-- 创建数据库和表结构

-- 创建数据库（如果不存在）
CREATE DATABASE IF NOT EXISTS cyber_inspector 
    DEFAULT CHARACTER SET utf8mb4 
    COLLATE utf8mb4_unicode_ci;

USE cyber_inspector;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) UNIQUE NOT NULL COMMENT '用户名',
    password VARCHAR(255) NOT NULL COMMENT '密码哈希',
    role ENUM('admin', 'user') DEFAULT 'user' COMMENT '用户角色',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    last_login DATETIME COMMENT '最后登录时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_username (username),
    INDEX idx_role (role),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- Agent 节点表
CREATE TABLE IF NOT EXISTS agents (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(64) NOT NULL COMMENT '节点名称',
    ip VARCHAR(15) NOT NULL COMMENT 'IP地址',
    url VARCHAR(128) UNIQUE NOT NULL COMMENT 'Agent URL',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    check_interval INT DEFAULT 300 COMMENT '巡检间隔（秒）',
    api_key VARCHAR(255) COMMENT 'API密钥',
    last_check_at DATETIME COMMENT '最后巡检时间',
    status ENUM('online', 'offline', 'unknown') DEFAULT 'unknown' COMMENT '节点状态',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_ip (ip),
    INDEX idx_enabled (enabled),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent节点表';

-- 巡检记录表
CREATE TABLE IF NOT EXISTS inspections (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    agent_id BIGINT UNSIGNED NOT NULL COMMENT 'Agent ID',
    hostname VARCHAR(64) NOT NULL COMMENT '主机名',
    ip VARCHAR(15) NOT NULL COMMENT 'IP地址',
    raw_data TEXT COMMENT '原始数据',
    analysis TEXT COMMENT '分析结果',
    alert BOOLEAN DEFAULT FALSE COMMENT '是否告警',
    level ENUM('OK', 'WARNING', 'CRITICAL') DEFAULT 'OK' COMMENT '告警级别',
    cpu_used DECIMAL(5,2) COMMENT 'CPU使用率',
    memory_used DECIMAL(5,2) COMMENT '内存使用率',
    disk_used DECIMAL(5,2) COMMENT '磁盘使用率',
    load_avg DECIMAL(5,2) COMMENT '平均负载',
    ping_loss DECIMAL(5,2) COMMENT '网络丢包率',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_agent_id (agent_id),
    INDEX idx_level (level),
    INDEX idx_alert (alert),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检记录表';

-- 告警记录表
CREATE TABLE IF NOT EXISTS alerts (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    agent_id BIGINT UNSIGNED NOT NULL COMMENT 'Agent ID',
    inspection_id BIGINT UNSIGNED NOT NULL COMMENT '巡检记录ID',
    level ENUM('WARNING', 'CRITICAL') NOT NULL COMMENT '告警级别',
    title VARCHAR(255) NOT NULL COMMENT '告警标题',
    summary TEXT COMMENT '告警摘要',
    details TEXT COMMENT '详细信息',
    solution TEXT COMMENT '解决方案',
    status ENUM('pending', 'processing', 'resolved', 'ignored') DEFAULT 'pending' COMMENT '告警状态',
    notified BOOLEAN DEFAULT FALSE COMMENT '是否已通知',
    resolved_at DATETIME COMMENT '解决时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_agent_id (agent_id),
    INDEX idx_level (level),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    FOREIGN KEY (inspection_id) REFERENCES inspections(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='告警记录表';

-- 登录日志表
CREATE TABLE IF NOT EXISTS login_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL COMMENT '用户名',
    ip VARCHAR(15) NOT NULL COMMENT '登录IP',
    status ENUM('success', 'failed') NOT NULL COMMENT '登录状态',
    message VARCHAR(255) COMMENT '登录消息',
    user_agent TEXT COMMENT 'User Agent',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_username (username),
    INDEX idx_ip (ip),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志表';

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id INT PRIMARY KEY AUTO_INCREMENT,
    config_key VARCHAR(64) UNIQUE NOT NULL COMMENT '配置键',
    config_value TEXT COMMENT '配置值',
    description VARCHAR(255) COMMENT '配置说明',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置表';

-- 初始化默认管理员用户（密码: admin123，请首次登录后修改）
INSERT IGNORE INTO users (username, password, role, enabled) VALUES 
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', TRUE);

-- 初始化系统配置
INSERT IGNORE INTO system_configs (config_key, config_value, description) VALUES 
('site_name', 'Cyber Inspector', '站点名称'),
('site_version', '2.0.0', '系统版本'),
('check_interval', '300', '默认巡检间隔（秒）'),
('alert_enabled', '1', '是否启用告警'),
('alert_cooldown', '300', '告警冷却时间（秒）'),
('mail_enabled', '0', '是否启用邮件通知'),
('theme', 'dark', '默认主题');

-- 创建视图：最新巡检记录
CREATE OR REPLACE VIEW latest_inspections AS
SELECT i.* 
FROM inspections i
INNER JOIN (
    SELECT agent_id, MAX(created_at) AS max_time 
    FROM inspections 
    GROUP BY agent_id
) t ON i.agent_id = t.agent_id AND i.created_at = t.max_time;

-- 创建视图：节点统计
CREATE OR REPLACE VIEW agent_stats AS
SELECT 
    a.id,
    a.name,
    a.ip,
    a.enabled,
    a.status,
    a.last_check_at,
    COALESCE(i.level, 'unknown') AS last_level,
    COALESCE(i.cpu_used, 0) AS cpu_used,
    COALESCE(i.memory_used, 0) AS memory_used,
    COALESCE(i.disk_used, 0) AS disk_used,
    COALESCE(i.created_at, a.created_at) AS last_inspection_at
FROM agents a
LEFT JOIN latest_inspections i ON a.id = i.agent_id;

-- 创建视图：告警统计
CREATE OR REPLACE VIEW alert_stats AS
SELECT 
    DATE(created_at) AS date,
    level,
    COUNT(*) AS count
FROM alerts
WHERE created_at >= DATE_SUB(CURDATE(), INTERVAL 30 DAY)
GROUP BY DATE(created_at), level
ORDER BY date DESC;

-- 权限设置（可选，根据实际需求调整）
-- GRANT SELECT, INSERT, UPDATE, DELETE ON cyber_inspector.* TO 'inspector'@'%';
-- FLUSH PRIVILEGES;

log_success "数据库初始化完成";
