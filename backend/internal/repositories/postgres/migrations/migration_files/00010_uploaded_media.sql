CREATE TYPE uploaded_media_mime_type AS ENUM ('image/png', 'image/jpeg', 'image/gif', 'video/mp4');

CREATE TABLE IF NOT EXISTS uploaded_media (
    id TEXT NOT NULL PRIMARY KEY,
    storage_path TEXT NOT NULL,
    mime_type uploaded_media_mime_type NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_updated_at TIMESTAMP WITH TIME ZONE,
    archived_at TIMESTAMP WITH TIME ZONE,
    created_by_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE
);
