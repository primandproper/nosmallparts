-- Data Privacy Domain Migration
-- GDPR/CCPA compliance tracking for user data disclosures

CREATE TYPE user_data_disclosure_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'expired'
);

CREATE TABLE IF NOT EXISTS user_data_disclosures (
    id TEXT NOT NULL PRIMARY KEY,
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    status user_data_disclosure_status NOT NULL DEFAULT 'pending',
    report_id TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    last_updated_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- INDEXES FOR DATA PRIVACY TABLES
-- =============================================================================

-- User data disclosures indexes
CREATE INDEX idx_user_data_disclosures_user ON user_data_disclosures (belongs_to_user);
CREATE INDEX idx_user_data_disclosures_status ON user_data_disclosures (status);
CREATE INDEX idx_user_data_disclosures_user_status ON user_data_disclosures (belongs_to_user, status);
CREATE INDEX idx_user_data_disclosures_expires_at ON user_data_disclosures (expires_at);
CREATE INDEX idx_user_data_disclosures_archived_at ON user_data_disclosures (archived_at);
