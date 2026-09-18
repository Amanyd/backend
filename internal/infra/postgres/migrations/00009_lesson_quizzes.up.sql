ALTER TABLE quizzes ADD COLUMN lesson_id UUID REFERENCES lessons(id) ON DELETE CASCADE;

ALTER TABLE quizzes DROP CONSTRAINT quizzes_course_id_difficulty_key;
CREATE UNIQUE INDEX idx_quiz_course_diff ON quizzes (course_id, difficulty) WHERE lesson_id IS NULL;
CREATE UNIQUE INDEX idx_quiz_lesson_diff ON quizzes (lesson_id, difficulty) WHERE lesson_id IS NOT NULL;

ALTER TABLE lessons ADD COLUMN keywords JSONB;
