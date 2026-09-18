ALTER TABLE lessons DROP COLUMN keywords;

DROP INDEX IF EXISTS idx_quiz_lesson_diff;
DROP INDEX IF EXISTS idx_quiz_course_diff;

ALTER TABLE quizzes ADD CONSTRAINT quizzes_course_id_difficulty_key UNIQUE(course_id, difficulty);
ALTER TABLE quizzes DROP COLUMN lesson_id;
