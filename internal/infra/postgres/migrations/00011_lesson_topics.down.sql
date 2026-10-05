ALTER TABLE questions DROP COLUMN IF EXISTS topic_phrase;
ALTER TABLE questions DROP COLUMN IF EXISTS difficulty;
ALTER TABLE questions DROP COLUMN IF EXISTS explanation;
ALTER TABLE questions DROP COLUMN IF EXISTS topic_id;

DROP TABLE IF EXISTS lesson_topics;
