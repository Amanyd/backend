CREATE TABLE files (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id      UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    file_name      TEXT NOT NULL,
    file_type      TEXT NOT NULL CHECK (file_type IN ('pdf', 'ppt', 'docx')),
    minio_key      TEXT NOT NULL,
    ingest_status  TEXT NOT NULL DEFAULT 'pending'
                   CHECK (ingest_status IN ('pending', 'processing', 'ready', 'failed')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_files_lesson ON files(lesson_id);
CREATE INDEX idx_files_status ON files(ingest_status);
