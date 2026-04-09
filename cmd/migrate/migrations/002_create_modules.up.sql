CREATE TABLE modules (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  practice_id   UUID NOT NULL REFERENCES practices(id) ON DELETE CASCADE,
  name         TEXT NOT NULL,  -- 'Reading and Writing 1', 'Math 2', etc.
  order_index  SMALLINT NOT NULL
);
