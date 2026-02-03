-- Settings Domain Migration
-- Service settings and configurations

CREATE TYPE setting_type AS ENUM (
    'user',
    'account',
    'membership'
);

CREATE TABLE IF NOT EXISTS service_settings (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT DEFAULT ''::TEXT NOT NULL,
    type setting_type DEFAULT 'user'::setting_type NOT NULL,
    description TEXT DEFAULT ''::TEXT NOT NULL,
    default_value TEXT,
    enumeration TEXT DEFAULT ''::TEXT NOT NULL,
    admins_only BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(name, archived_at)
);

CREATE TABLE IF NOT EXISTS service_setting_configurations (
    id TEXT NOT NULL PRIMARY KEY,
    value TEXT DEFAULT ''::TEXT NOT NULL,
    notes TEXT DEFAULT ''::TEXT NOT NULL,
    service_setting_id TEXT NOT NULL REFERENCES service_settings("id") ON DELETE CASCADE,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    belongs_to_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(belongs_to_user, belongs_to_account, service_setting_id)
);

-- =============================================================================
-- INDEXES FOR SETTINGS TABLES
-- =============================================================================

-- Service settings indexes
CREATE INDEX idx_service_settings_archived_at ON service_settings (archived_at) WHERE archived_at IS NULL;
CREATE INDEX idx_service_settings_name ON service_settings (name) WHERE archived_at IS NULL;
CREATE INDEX idx_service_settings_type ON service_settings (type) WHERE archived_at IS NULL;
CREATE INDEX idx_service_settings_admins_only ON service_settings (admins_only) WHERE archived_at IS NULL;

-- Service setting configurations indexes
CREATE INDEX idx_setting_configs_user ON service_setting_configurations (belongs_to_user) WHERE archived_at IS NULL;
CREATE INDEX idx_setting_configs_account ON service_setting_configurations (belongs_to_account) WHERE archived_at IS NULL;
CREATE INDEX idx_setting_configs_setting ON service_setting_configurations (service_setting_id) WHERE archived_at IS NULL;
