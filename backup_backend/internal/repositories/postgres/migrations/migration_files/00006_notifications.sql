-- Notifications Domain Migration
-- User notification management

CREATE TYPE user_notification_status AS ENUM (
    'unread',
    'read',
    'dismissed'
);

CREATE TABLE IF NOT EXISTS user_notifications (
    id TEXT NOT NULL PRIMARY KEY,
    content TEXT NOT NULL,
    status user_notification_status NOT NULL DEFAULT 'unread',
    belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    last_updated_at TIMESTAMP WITH TIME ZONE
);

-- =============================================================================
-- INDEXES FOR NOTIFICATIONS TABLES
-- =============================================================================

-- User notifications indexes
CREATE INDEX idx_user_notifications_user ON user_notifications (belongs_to_user);
CREATE INDEX idx_user_notifications_status ON user_notifications (status);
CREATE INDEX idx_user_notifications_user_status ON user_notifications (belongs_to_user, status);
CREATE INDEX idx_user_notifications_user_created_at ON user_notifications (belongs_to_user, created_at);
