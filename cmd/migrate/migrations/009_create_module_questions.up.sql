CREATE TABLE module_questions (
  module_id    UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
  question_id  UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
  number       SMALLINT NOT NULL,
  PRIMARY KEY (module_id, question_id)
);

