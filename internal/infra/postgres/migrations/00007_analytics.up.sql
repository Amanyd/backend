CREATE TABLE events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id  UUID REFERENCES courses(id) ON DELETE SET NULL,
    type       TEXT NOT NULL,
    metadata   JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_user ON events(user_id);
CREATE INDEX idx_events_course ON events(course_id);
CREATE INDEX idx_events_type ON events(type);
