CREATE TABLE IF NOT EXISTS waitlists (
    id TEXT NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS waitlist_signups (
    id TEXT NOT NULL PRIMARY KEY,
    notes TEXT NOT NULL DEFAULT '',
    belongs_to_waitlist TEXT REFERENCES waitlists("id") ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    belongs_to_user TEXT REFERENCES users("id") ON DELETE CASCADE,
    belongs_to_account TEXT REFERENCES accounts("id") ON DELETE CASCADE
);
