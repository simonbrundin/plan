-- Preview/demo data for ephemeral PR environments

INSERT INTO users (sub, email, first_name, last_name) VALUES
  ('preview-user-1', 'preview@test.local', 'Preview', 'User');

INSERT INTO goals (title, finished) VALUES
  ('Sample Goal 1', NULL),
  ('Sample Goal 2', CURRENT_TIMESTAMP),
  ('Sample Goal 3', NULL);

INSERT INTO user_goals (user_id, goal_id) VALUES
  (1, 1), (1, 2), (1, 3);

INSERT INTO goal_relations (parent_id, child_id) VALUES
  (1, 2), (1, 3);