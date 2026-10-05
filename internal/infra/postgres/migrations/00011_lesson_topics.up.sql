CREATE TABLE lesson_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    order_index INT NOT NULL DEFAULT 0,
    slides JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_lesson_topics_lesson_id ON lesson_topics(lesson_id);

ALTER TABLE questions ADD COLUMN IF NOT EXISTS topic_id UUID REFERENCES lesson_topics(id) ON DELETE SET NULL;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS explanation TEXT;
ALTER TABLE questions ADD COLUMN IF NOT EXISTS difficulty TEXT DEFAULT 'medium';
ALTER TABLE questions ADD COLUMN IF NOT EXISTS topic_phrase TEXT;
