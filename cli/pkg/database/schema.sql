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