CREATE TABLE IF NOT EXISTS friends(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user1 UUID  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user2 UUID  NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE friends ADD CONSTRAINT unique_friendship 
  UNIQUE (user1, user2);

ALTER TABLE friends ADD CONSTRAINT enforce_user_order 
  CHECK (user1 < user2);

ALTER TABLE friends ADD CONSTRAINT no_self_friendship 
  CHECK (user1 <> user2);
