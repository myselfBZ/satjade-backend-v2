CREATE TABLE friendship_requests(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    to_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    message TEXT
);


ALTER TABLE friendship_requests ADD CONSTRAINT unique_request 
  UNIQUE (from_id, to_id);

ALTER TABLE friendship_requests ADD CONSTRAINT no_self_request 
  CHECK (from_id <> to_id);
