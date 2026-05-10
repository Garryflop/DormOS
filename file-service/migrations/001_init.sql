-- Files metadata table
CREATE TABLE IF NOT EXISTS files (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename     VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    bucket       VARCHAR(100) NOT NULL DEFAULT 'dormos-files',
    object_key   VARCHAR(512) NOT NULL UNIQUE,
    uploaded_by  UUID         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by);
