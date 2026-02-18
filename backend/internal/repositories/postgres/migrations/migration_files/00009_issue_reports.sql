CREATE TABLE IF NOT EXISTS issue_reports (
    id TEXT NOT NULL PRIMARY KEY,
    issue_type TEXT NOT NULL,
    details TEXT NOT NULL,
    relevant_table TEXT,
    relevant_record_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    created_by_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE,
    belongs_to_account TEXT NOT NULL REFERENCES accounts("id") ON DELETE CASCADE
);
