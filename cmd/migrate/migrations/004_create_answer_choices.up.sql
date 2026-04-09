CREATE TABLE answer_choices (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  label        CHAR(1) NOT NULL,  -- 'A', 'B', 'C', 'D'
  body         TEXT NOT NULL,
  UNIQUE (question_id, label)
);

ALTER TABLE questions
  ADD CONSTRAINT fk_correct_choice
  FOREIGN KEY (correct_choice_id)
  REFERENCES answer_choices(id)
  DEFERRABLE INITIALLY DEFERRED;
