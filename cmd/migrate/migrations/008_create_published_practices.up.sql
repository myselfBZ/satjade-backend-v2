CREATE TABLE published_practices (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  practice_id  UUID NOT NULL REFERENCES practices(id),
  published_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  published_by UUID NOT NULL REFERENCES users(id), 
  data         JSONB NOT NULL
);
