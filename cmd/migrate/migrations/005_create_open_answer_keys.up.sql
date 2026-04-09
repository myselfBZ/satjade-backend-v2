CREATE TABLE open_answer_keys (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id    UUID NOT NULL UNIQUE REFERENCES questions(id) ON DELETE CASCADE,
  model_answer   TEXT,     -- human-readable expected answer
  match_pattern  TEXT      -- optional regex for auto-grading
);
