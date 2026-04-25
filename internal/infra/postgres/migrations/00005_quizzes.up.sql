CREATE TABLE quizzes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    difficulty  TEXT NOT NULL CHECK (difficulty IN ('easy', 'medium', 'hard')),
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'generating', 'ready', 'failed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(course_id, difficulty)
);

CREATE TABLE questions (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id   UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    type      TEXT NOT NULL CHECK (type IN ('mcq', 'open_ended')),
    question  TEXT NOT NULL,
    choices   JSONB,
    answer    TEXT NOT NULL,
    order_idx INT NOT NULL DEFAULT 0
);

CREATE TABLE attempts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id    UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score      DOUBLE PRECISION NOT NULL DEFAULT 0,
    total      INT NOT NULL DEFAULT 0,
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at   TIMESTAMPTZ
);

CREATE TABLE answers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id  UUID NOT NULL REFERENCES attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    user_answer TEXT NOT NULL,
    is_correct  BOOLEAN NOT NULL DEFAULT false
);
