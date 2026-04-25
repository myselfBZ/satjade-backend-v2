CREATE TABLE attempt_responses (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  attempt_id         UUID NOT NULL REFERENCES practice_attempts(id) ON DELETE CASCADE,
  question_id        UUID NOT NULL REFERENCES questions(id),
  -- MC: reference to a choice. Open: NULL
  selected_choice_id UUID REFERENCES answer_choices(id),
  -- Open response: raw text. MC: NULL
  open_response      TEXT,
  is_correct         BOOLEAN, 
  UNIQUE (attempt_id, question_id),

  CONSTRAINT chk_response_type CHECK (
    (selected_choice_id IS NOT NULL AND open_response IS NULL)
    OR
    (selected_choice_id IS NULL     AND open_response IS NOT NULL)
    OR
    (selected_choice_id IS NULL     AND open_response IS NULL)  -- unanswered
  )
);


CREATE INDEX idx_responses_attempt  ON attempt_responses(attempt_id);
