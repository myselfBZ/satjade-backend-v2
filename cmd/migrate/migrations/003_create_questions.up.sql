CREATE TYPE question_type AS ENUM ('multiple_choice', 'open_response');
CREATE TYPE difficulty     AS ENUM ('easy', 'medium', 'hard');

CREATE TABLE questions (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  type               question_type NOT NULL,
  paragraph          TEXT,          -- NULL for Math
  prompt             TEXT NOT NULL,
  image_path         TEXT,          -- single image, nullable
  skill              TEXT,
  domain             TEXT,
  difficulty         difficulty NOT NULL,
  explanation        TEXT,
  -- MC only: points to the correct choice. 
  correct_choice_id  UUID

  -- couldnt get it working
  -- Enforce: open_response questions must NOT have a correct_choice_id
  -- CONSTRAINT chk_answer_type CHECK (
  --   (type = 'multiple_choice' AND correct_choice_id IS NOT NULL)
  --   OR
  --   (type = 'open_response'  AND correct_choice_id IS NULL)
  -- )
);



