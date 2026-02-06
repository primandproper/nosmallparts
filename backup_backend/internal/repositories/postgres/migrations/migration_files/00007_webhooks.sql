-- Webhooks Domain Migration
-- Webhook management and trigger events

CREATE TYPE webhook_event AS ENUM (
    'webhook_created',
    'webhook_updated',
    'webhook_archived'
);

CREATE TABLE IF NOT EXISTS webhooks (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    url TEXT NOT NULL,
    method TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    belongs_to_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS webhook_trigger_events (
    id TEXT NOT NULL PRIMARY KEY,
    trigger_event webhook_event NOT NULL,
    belongs_to_webhook TEXT NOT NULL REFERENCES webhooks("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(trigger_event, belongs_to_webhook)
);

-- =============================================================================
-- INDEXES FOR WEBHOOKS TABLES
-- =============================================================================

-- Webhooks indexes
CREATE INDEX idx_webhooks_account ON webhooks (belongs_to_account) WHERE archived_at IS NULL;
CREATE INDEX idx_webhooks_archived_at ON webhooks (archived_at) WHERE archived_at IS NULL;

-- Webhook trigger events indexes
CREATE INDEX idx_webhook_triggers_webhook ON webhook_trigger_events (belongs_to_webhook) WHERE archived_at IS NULL;
CREATE INDEX idx_webhook_triggers_event ON webhook_trigger_events (trigger_event) WHERE archived_at IS NULL;
