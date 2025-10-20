-- scripts/init_db.sql
-- TCP代理桥接系统数据库初始化脚本
-- 注意：移除了默认服务器数据插入，由应用自动从配置文件同步

-- =============================================
-- 消息队列表：存储所有接收到的TCP消息
-- =============================================
CREATE TABLE IF NOT EXISTS message_queue (
    id BIGSERIAL PRIMARY KEY,                          -- 消息ID，自增主键
    source_ip INET NOT NULL,                           -- 源IP地址
    original_data BYTEA NOT NULL,                      -- 原始消息数据（二进制格式）
    data_length INTEGER NOT NULL,                      -- 数据长度（字节数）
    created_at TIMESTAMP DEFAULT NOW(),                -- 消息创建时间
    processed_at TIMESTAMP NULL,                       -- 消息处理完成时间
    status VARCHAR(20) DEFAULT 'received'              -- 消息状态: received-已接收
);

-- 表注释
COMMENT ON TABLE message_queue IS 'TCP消息队列表，存储所有接收到的消息数据';

-- 字段注释
COMMENT ON COLUMN message_queue.id IS '消息唯一标识，自增主键';
COMMENT ON COLUMN message_queue.source_ip IS '消息来源IP地址';
COMMENT ON COLUMN message_queue.original_data IS '原始消息二进制数据';
COMMENT ON COLUMN message_queue.data_length IS '消息数据长度（字节）';
COMMENT ON COLUMN message_queue.created_at IS '消息创建时间';
COMMENT ON COLUMN message_queue.processed_at IS '消息处理完成时间';
COMMENT ON COLUMN message_queue.status IS '消息状态: received-已接收';

-- =============================================
-- 目标投递状态表：记录每个消息到每个目标服务器的投递状态
-- =============================================
CREATE TABLE IF NOT EXISTS target_delivery_status (
    id BIGSERIAL PRIMARY KEY,                          -- 投递状态ID
    message_id BIGINT NOT NULL REFERENCES message_queue(id) ON DELETE CASCADE, -- 关联消息ID
    target_server_id VARCHAR(50) NOT NULL,             -- 目标服务器ID
    target_server_name VARCHAR(100) NOT NULL,          -- 目标服务器名称
    target_address VARCHAR(100) NOT NULL,              -- 目标服务器地址
    
    -- 发送状态相关字段
    status VARCHAR(20) DEFAULT 'pending',              -- 投递状态: pending-等待发送, sending-发送中, sent-已发送, failed-发送失败
    send_attempts INTEGER DEFAULT 0,                   -- 发送尝试次数
    max_attempts INTEGER DEFAULT 5,                    -- 最大尝试次数
    last_attempt_at TIMESTAMP NULL,                    -- 最后尝试时间
    next_retry_at TIMESTAMP NULL,                      -- 下次重试时间
    sent_at TIMESTAMP NULL,                            -- 成功发送时间
    
    -- 错误处理相关字段
    last_error TEXT NULL,                              -- 最后错误信息
    data_size INTEGER NOT NULL,                        -- 数据大小
    
    -- 时间戳字段
    created_at TIMESTAMP DEFAULT NOW(),                -- 创建时间
    updated_at TIMESTAMP DEFAULT NOW()                 -- 更新时间
);

-- 表注释
COMMENT ON TABLE target_delivery_status IS '目标服务器投递状态表，记录每个消息的投递情况';

-- 字段注释
COMMENT ON COLUMN target_delivery_status.id IS '投递状态唯一标识';
COMMENT ON COLUMN target_delivery_status.message_id IS '关联的消息ID';
COMMENT ON COLUMN target_delivery_status.target_server_id IS '目标服务器ID';
COMMENT ON COLUMN target_delivery_status.target_server_name IS '目标服务器名称';
COMMENT ON COLUMN target_delivery_status.target_address IS '目标服务器地址';
COMMENT ON COLUMN target_delivery_status.status IS '投递状态: pending-等待发送, sending-发送中, sent-已发送, failed-发送失败';
COMMENT ON COLUMN target_delivery_status.send_attempts IS '已尝试发送次数';
COMMENT ON COLUMN target_delivery_status.max_attempts IS '最大允许尝试次数';
COMMENT ON COLUMN target_delivery_status.last_attempt_at IS '最后一次尝试发送时间';
COMMENT ON COLUMN target_delivery_status.next_retry_at IS '下次重试时间';
COMMENT ON COLUMN target_delivery_status.sent_at IS '成功发送时间';
COMMENT ON COLUMN target_delivery_status.last_error IS '最后一次错误信息';
COMMENT ON COLUMN target_delivery_status.data_size IS '消息数据大小';
COMMENT ON COLUMN target_delivery_status.created_at IS '记录创建时间';
COMMENT ON COLUMN target_delivery_status.updated_at IS '记录最后更新时间';

-- =============================================
-- 目标服务器配置表：存储目标服务器信息
-- =============================================
CREATE TABLE IF NOT EXISTS target_servers (
    id VARCHAR(50) PRIMARY KEY,                        -- 服务器ID
    name VARCHAR(100) NOT NULL,                        -- 服务器名称
    address VARCHAR(100) NOT NULL,                     -- 服务器地址
    enabled BOOLEAN DEFAULT true,                      -- 是否启用
    is_online BOOLEAN DEFAULT false,                   -- 是否在线
    last_health_check TIMESTAMP NULL,                  -- 最后健康检查时间
    
    -- 连接配置
    connection_timeout_sec INTEGER DEFAULT 10,         -- 连接超时时间（秒）
    max_retries INTEGER DEFAULT 5,                     -- 最大重试次数
    batch_size INTEGER DEFAULT 100,                    -- 批量处理大小
    priority INTEGER DEFAULT 5,                        -- 优先级（数字越小优先级越高）
    
    -- 统计信息
    total_messages_sent BIGINT DEFAULT 0,              -- 总发送消息数
    total_errors BIGINT DEFAULT 0,                     -- 总错误数
    last_success_at TIMESTAMP NULL,                    -- 最后成功时间
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT NOW(),                -- 创建时间
    updated_at TIMESTAMP DEFAULT NOW()                 -- 更新时间
);

-- 表注释
COMMENT ON TABLE target_servers IS '目标服务器配置表';

-- 字段注释
COMMENT ON COLUMN target_servers.id IS '服务器唯一标识';
COMMENT ON COLUMN target_servers.name IS '服务器名称';
COMMENT ON COLUMN target_servers.address IS '服务器地址（IP:Port）';
COMMENT ON COLUMN target_servers.enabled IS '是否启用该服务器';
COMMENT ON COLUMN target_servers.is_online IS '服务器当前是否在线';
COMMENT ON COLUMN target_servers.last_health_check IS '最后一次健康检查时间';
COMMENT ON COLUMN target_servers.connection_timeout_sec IS '连接超时时间（秒）';
COMMENT ON COLUMN target_servers.max_retries IS '消息发送最大重试次数';
COMMENT ON COLUMN target_servers.batch_size IS '批量处理消息数量';
COMMENT ON COLUMN target_servers.priority IS '服务器优先级（数字越小优先级越高）';
COMMENT ON COLUMN target_servers.total_messages_sent IS '成功发送的消息总数';
COMMENT ON COLUMN target_servers.total_errors IS '发送错误总数';
COMMENT ON COLUMN target_servers.last_success_at IS '最后一次成功发送时间';
COMMENT ON COLUMN target_servers.created_at IS '记录创建时间';
COMMENT ON COLUMN target_servers.updated_at IS '记录最后更新时间';

-- =============================================
-- 性能优化索引
-- =============================================

-- 消息队列表索引
CREATE INDEX IF NOT EXISTS idx_message_queue_status ON message_queue(status, created_at);
COMMENT ON INDEX idx_message_queue_status IS '消息状态索引，用于快速查询待处理消息';

CREATE INDEX IF NOT EXISTS idx_message_queue_created_at ON message_queue(created_at);
COMMENT ON INDEX idx_message_queue_created_at IS '消息创建时间索引，用于按时间范围查询';

CREATE INDEX IF NOT EXISTS idx_message_queue_source_ip ON message_queue(source_ip);
COMMENT ON INDEX idx_message_queue_source_ip IS '消息来源IP索引，用于按来源查询';

-- 目标投递状态表索引
CREATE INDEX IF NOT EXISTS idx_delivery_status_message ON target_delivery_status(message_id);
COMMENT ON INDEX idx_delivery_status_message IS '投递状态消息ID索引，用于关联查询';

CREATE INDEX IF NOT EXISTS idx_delivery_status_target ON target_delivery_status(target_server_id);
COMMENT ON INDEX idx_delivery_status_target IS '投递状态目标服务器ID索引，用于按目标服务器查询';

CREATE INDEX IF NOT EXISTS idx_delivery_status_composite ON target_delivery_status(status, next_retry_at, send_attempts);
COMMENT ON INDEX idx_delivery_status_composite IS '投递状态复合索引，用于快速查询需要重试的消息';

CREATE INDEX IF NOT EXISTS idx_delivery_status_retry ON target_delivery_status(next_retry_at) 
WHERE status IN ('pending', 'failed');
COMMENT ON INDEX idx_delivery_status_retry IS '重试时间条件索引，用于快速查询可重试的消息';

CREATE INDEX IF NOT EXISTS idx_delivery_status_sent_at ON target_delivery_status(sent_at);
COMMENT ON INDEX idx_delivery_status_sent_at IS '发送时间索引，用于统计和清理';

-- 目标服务器表索引
CREATE INDEX IF NOT EXISTS idx_target_servers_online ON target_servers(enabled, is_online);
COMMENT ON INDEX idx_target_servers_online IS '目标服务器在线状态索引，用于快速查询可用服务器';

CREATE INDEX IF NOT EXISTS idx_target_servers_priority ON target_servers(priority);
COMMENT ON INDEX idx_target_servers_priority IS '目标服务器优先级索引，用于按优先级排序';

CREATE INDEX IF NOT EXISTS idx_target_servers_address ON target_servers(address);
COMMENT ON INDEX idx_target_servers_address IS '目标服务器地址索引，用于快速查找';

-- =============================================
-- 自动维护函数和触发器
-- =============================================

-- 更新时间戳自动更新函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS '自动更新updated_at字段的触发器函数';

-- 为目标服务器表创建更新时间触发器
CREATE TRIGGER update_target_servers_updated_at
    BEFORE UPDATE ON target_servers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TRIGGER update_target_servers_updated_at ON target_servers IS '自动更新目标服务器表更新时间';

-- 为投递状态表创建更新时间触发器
CREATE TRIGGER update_delivery_status_updated_at
    BEFORE UPDATE ON target_delivery_status
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TRIGGER update_delivery_status_updated_at ON target_delivery_status IS '自动更新投递状态表更新时间';

-- =============================================
-- 数据清理和维护函数
-- =============================================

-- 清理旧消息的函数（保留最近30天的数据）
CREATE OR REPLACE FUNCTION cleanup_old_messages()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- 删除30天前已发送或失败的消息
    WITH deleted AS (
        DELETE FROM message_queue 
        WHERE created_at < NOW() - INTERVAL '30 days'
        AND status IN ('sent', 'failed')
        RETURNING id
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_messages() IS '清理30天前的已发送或失败消息，返回删除的记录数';

-- 清理关联的投递状态数据（级联删除）
CREATE OR REPLACE FUNCTION cleanup_old_delivery_status()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- 删除30天前已发送或失败的投递状态记录
    WITH deleted AS (
        DELETE FROM target_delivery_status 
        WHERE created_at < NOW() - INTERVAL '30 days'
        AND status IN ('sent', 'failed')
        RETURNING id
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION cleanup_old_delivery_status() IS '清理30天前的已发送或失败投递状态记录，返回删除的记录数';

-- 获取系统统计信息的函数
CREATE OR REPLACE FUNCTION get_system_stats()
RETURNS TABLE(
    total_messages BIGINT,
    pending_messages BIGINT,
    sent_messages BIGINT,
    failed_messages BIGINT,
    total_servers BIGINT,
    online_servers BIGINT
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        (SELECT COUNT(*) FROM message_queue) as total_messages,
        (SELECT COUNT(*) FROM message_queue WHERE status = 'received') as pending_messages,
        (SELECT COUNT(*) FROM target_delivery_status WHERE status = 'sent') as sent_messages,
        (SELECT COUNT(*) FROM target_delivery_status WHERE status = 'failed') as failed_messages,
        (SELECT COUNT(*) FROM target_servers WHERE enabled = true) as total_servers,
        (SELECT COUNT(*) FROM target_servers WHERE enabled = true AND is_online = true) as online_servers;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION get_system_stats() IS '获取系统统计信息，包括消息数量和服务器状态';

-- =============================================
-- 定期维护任务建议
-- =============================================
-- 建议设置以下定期任务（使用pg_cron或其他调度工具）：
-- 
-- 1. 每天清理旧数据：
--    SELECT cleanup_old_messages();
--    SELECT cleanup_old_delivery_status();
--
-- 2. 每小时更新服务器在线状态：
--    UPDATE target_servers SET is_online = false 
--    WHERE enabled = true AND last_health_check < NOW() - INTERVAL '5 minutes';
--
-- 3. 每周重新构建索引：
--    REINDEX TABLE message_queue;
--    REINDEX TABLE target_delivery_status;
--    REINDEX TABLE target_servers;

-- =============================================
-- 数据库初始化完成提示
-- =============================================
DO $$
BEGIN
    RAISE NOTICE 'TCP代理桥接系统数据库初始化完成';
    RAISE NOTICE '请确保在应用启动前配置configs/config.yaml文件中的目标服务器信息';
    RAISE NOTICE '应用启动时将自动同步配置到数据库';
END $$;