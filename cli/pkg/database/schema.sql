-- Schema for Shipyard deployment versioning database
-- SQLite database to track deployment history and versions

CREATE TABLE IF NOT EXISTS apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS deployments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    version TEXT NOT NULL,
    image TEXT NOT NULL,
    image_tag TEXT NOT NULL,
    image_hash TEXT NOT NULL,
    config_json TEXT NOT NULL, -- JSON serialized config
    config_hash TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'success', 'failed')),
    rollback_to_version TEXT, -- NULL if not a rollback
    deployed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME, -- When deployment finished (success or failed)
    error_message TEXT, -- Error details if deployment failed
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE (app_id, version)
);

-- Index for fast queries
CREATE INDEX IF NOT EXISTS idx_deployments_app_id ON deployments(app_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_deployed_at ON deployments(deployed_at DESC);
CREATE INDEX IF NOT EXISTS idx_deployments_app_status ON deployments(app_id, status);

-- View for easy querying with app names
CREATE VIEW IF NOT EXISTS deployment_history AS
SELECT 
    d.id,
    a.name as app_name,
    d.version,
    d.image,
    d.image_tag,
    d.image_hash,
    d.config_json,
    d.config_hash,
    d.status,
    d.rollback_to_version,
    d.deployed_at,
    d.completed_at,
    d.error_message
FROM deployments d
JOIN apps a ON d.app_id = a.id
ORDER BY d.deployed_at DESC;

-- Trigger to update apps.updated_at when new deployment is added
CREATE TRIGGER IF NOT EXISTS update_app_timestamp 
    AFTER INSERT ON deployments
BEGIN
    UPDATE apps SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.app_id;
END;

-- Table for domain management
CREATE TABLE IF NOT EXISTS domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    hostname TEXT NOT NULL,
    base_domain TEXT NOT NULL, -- extracted base domain (e.g., example.com)
    path TEXT DEFAULT '/', -- path prefix (for future use)
    ssl_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE,
    UNIQUE (hostname) -- Each hostname can only belong to one app
);

-- Index for fast domain queries
CREATE INDEX IF NOT EXISTS idx_domains_app_id ON domains(app_id);
CREATE INDEX IF NOT EXISTS idx_domains_hostname ON domains(hostname);
CREATE INDEX IF NOT EXISTS idx_domains_base_domain ON domains(base_domain);

-- View for domains with app names
CREATE VIEW IF NOT EXISTS domain_overview AS
SELECT 
    d.id,
    a.name as app_name,
    d.hostname,
    d.base_domain,
    d.path,
    d.ssl_enabled,
    d.created_at,
    d.updated_at
FROM domains d
JOIN apps a ON d.app_id = a.id
ORDER BY d.base_domain, d.hostname;

-- Trigger to update domains.updated_at
CREATE TRIGGER IF NOT EXISTS update_domain_timestamp 
    AFTER UPDATE ON domains
BEGIN
    UPDATE domains SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Table for container registry credentials
CREATE TABLE IF NOT EXISTS registry_credentials (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    registry_url TEXT NOT NULL UNIQUE, -- e.g., ghcr.io, docker.io, my-registry.com
    username TEXT NOT NULL,
    password TEXT NOT NULL, -- Encrypted token/password
    email TEXT, -- Optional email for Docker registry
    registry_type TEXT DEFAULT 'docker', -- docker, github, gitlab, aws, etc.
    is_default BOOLEAN DEFAULT FALSE, -- Default registry for pulling
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast registry queries
CREATE INDEX IF NOT EXISTS idx_registry_url ON registry_credentials(registry_url);
CREATE INDEX IF NOT EXISTS idx_registry_default ON registry_credentials(is_default);

-- Trigger to update registry_credentials.updated_at
CREATE TRIGGER IF NOT EXISTS update_registry_timestamp 
    AFTER UPDATE ON registry_credentials
BEGIN
    UPDATE registry_credentials SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Trigger to ensure only one default registry
CREATE TRIGGER IF NOT EXISTS ensure_single_default_registry
    AFTER UPDATE OF is_default ON registry_credentials
    WHEN NEW.is_default = 1
BEGIN
    UPDATE registry_credentials SET is_default = 0 WHERE id != NEW.id AND is_default = 1;
END;

CREATE TRIGGER IF NOT EXISTS ensure_single_default_registry_insert
    AFTER INSERT ON registry_credentials
    WHEN NEW.is_default = 1
BEGIN
    UPDATE registry_credentials SET is_default = 0 WHERE id != NEW.id AND is_default = 1;
END;

-- Tables for monitoring and surveillance
CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    metric_type TEXT NOT NULL, -- cpu, memory, network, disk, pods, requests
    value REAL NOT NULL,
    unit TEXT, -- percentage, bytes, count, ms
    pod_name TEXT, -- specific pod if applicable
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS health_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    endpoint TEXT NOT NULL, -- /health, /ready, etc.
    method TEXT DEFAULT 'GET',
    status TEXT NOT NULL, -- healthy, unhealthy, timeout, error
    status_code INTEGER, -- HTTP status code
    response_time INTEGER, -- response time in ms
    error_message TEXT,
    checked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL,
    alert_type TEXT NOT NULL, -- cpu_high, memory_high, error_rate, response_time
    threshold REAL NOT NULL,
    current_value REAL NOT NULL,
    severity TEXT NOT NULL, -- info, warning, critical
    status TEXT NOT NULL, -- active, resolved, suppressed
    message TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    acknowledged_at DATETIME,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER, -- NULL for cluster-wide events
    event_type TEXT NOT NULL, -- Normal, Warning
    reason TEXT NOT NULL, -- Started, Failed, FailedScheduling, etc.
    message TEXT NOT NULL,
    object_kind TEXT, -- Pod, Deployment, Service, etc.
    object_name TEXT,
    first_timestamp DATETIME,
    last_timestamp DATETIME,
    count INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

-- Monitoring configuration per app
CREATE TABLE IF NOT EXISTS monitoring_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    app_id INTEGER NOT NULL UNIQUE,
    enabled BOOLEAN DEFAULT TRUE,
    health_check_path TEXT DEFAULT '/health',
    health_check_interval INTEGER DEFAULT 30, -- seconds
    health_check_timeout INTEGER DEFAULT 5, -- seconds
    metrics_enabled BOOLEAN DEFAULT TRUE,
    metrics_path TEXT DEFAULT '/metrics',
    metrics_port INTEGER DEFAULT 9090,
    retention_days INTEGER DEFAULT 7,
    cpu_threshold REAL DEFAULT 80.0, -- percentage
    memory_threshold REAL DEFAULT 85.0, -- percentage
    error_rate_threshold REAL DEFAULT 5.0, -- percentage
    response_time_threshold INTEGER DEFAULT 1000, -- milliseconds
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (app_id) REFERENCES apps(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_metrics_app_timestamp ON metrics(app_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_type_timestamp ON metrics(metric_type, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_health_checks_app_timestamp ON health_checks(app_id, checked_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_app_status ON alerts(app_id, status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_app_timestamp ON events(app_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_events_type_timestamp ON events(event_type, created_at DESC);

-- Views for easy querying
CREATE VIEW IF NOT EXISTS latest_metrics AS
SELECT 
    a.name as app_name,
    m.metric_type,
    m.value,
    m.unit,
    m.pod_name,
    m.timestamp
FROM metrics m
JOIN apps a ON m.app_id = a.id
WHERE m.timestamp > datetime('now', '-1 hour')
ORDER BY m.timestamp DESC;

CREATE VIEW IF NOT EXISTS active_alerts AS
SELECT 
    a.name as app_name,
    al.alert_type,
    al.threshold,
    al.current_value,
    al.severity,
    al.message,
    al.created_at
FROM alerts al
JOIN apps a ON al.app_id = a.id
WHERE al.status = 'active'
ORDER BY 
    CASE al.severity 
        WHEN 'critical' THEN 1 
        WHEN 'warning' THEN 2 
        WHEN 'info' THEN 3 
    END,
    al.created_at DESC;

CREATE VIEW IF NOT EXISTS app_health_summary AS
SELECT 
    a.name as app_name,
    COUNT(CASE WHEN hc.status = 'healthy' THEN 1 END) as healthy_checks,
    COUNT(CASE WHEN hc.status != 'healthy' THEN 1 END) as unhealthy_checks,
    AVG(hc.response_time) as avg_response_time,
    MAX(hc.checked_at) as last_check
FROM apps a
LEFT JOIN health_checks hc ON a.id = hc.app_id 
    AND hc.checked_at > datetime('now', '-1 hour')
GROUP BY a.id, a.name;

-- Triggers for monitoring
CREATE TRIGGER IF NOT EXISTS update_monitoring_config_timestamp 
    AFTER UPDATE ON monitoring_config
BEGIN
    UPDATE monitoring_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Cleanup old data trigger (called periodically)
CREATE TRIGGER IF NOT EXISTS cleanup_old_metrics
    AFTER INSERT ON metrics
BEGIN
    DELETE FROM metrics WHERE timestamp < datetime('now', '-7 days');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_health_checks
    AFTER INSERT ON health_checks
BEGIN
    DELETE FROM health_checks WHERE checked_at < datetime('now', '-7 days');
END;

CREATE TRIGGER IF NOT EXISTS cleanup_old_events
    AFTER INSERT ON events
BEGIN
    DELETE FROM events WHERE created_at < datetime('now', '-30 days');
END;