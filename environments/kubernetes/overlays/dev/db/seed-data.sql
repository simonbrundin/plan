-- Development test data
-- Add your dev user and test goals here

INSERT INTO users (email, first_name, last_name) VALUES
  ('simon@dev.local', 'Simon', 'Dev'),
  ('alice@dev.local', 'Alice', 'Developer'),
  ('bob@dev.local', 'Bob', 'Tester');

INSERT INTO goals (title, finished) VALUES
  ('Build MVP', NULL),
  ('Setup development environment', CURRENT_TIMESTAMP),
  ('Write tests', NULL),
  ('Deploy to staging', NULL),
  ('User authentication', CURRENT_TIMESTAMP),
  ('API endpoints', NULL),
  ('Frontend components', NULL),
  ('Database schema', CURRENT_TIMESTAMP),
  ('CI/CD pipeline', NULL),
  ('Documentation', NULL);

INSERT INTO user_goals (user_id, goal_id) VALUES
  (1, 1), (1, 2), (1, 3), (1, 4), (1, 5),
  (2, 6), (2, 7), (2, 8),
  (3, 9), (3, 10);

INSERT INTO goal_relations (parent_id, child_id) VALUES
  (1, 2), (1, 5), (1, 6),
  (5, 7), (5, 8),
  (4, 9);
