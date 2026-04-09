CREATE TABLE practice_attempts (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  published_practice_id   UUID NOT NULL REFERENCES published_practices(id) ON DELETE CASCADE,
  started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  submitted_at  TIMESTAMPTZ,
  rw_score      SMALLINT NOT NULL, 
  math_score    SMALLINT NOT NULL 
);


CREATE INDEX idx_attempts_users  ON practice_attempts(user_id);
CREATE INDEX idx_attempts_practice ON practice_attempts(published_practice_id);

