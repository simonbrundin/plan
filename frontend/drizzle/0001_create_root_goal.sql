-- Migration: Create root goal with ID 1 and migrate parentless goals
-- This creates a root goal and makes all current parentless goals children of it

BEGIN;

-- Create root goal with ID 1 (if it doesn't exist)
INSERT INTO goals (id, title, created)
VALUES (1, 'Root', CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Create goal_relations for all goals that don't have parents
-- (goals that are not in goal_relations as child_id)
INSERT INTO goal_relations (parent_id, child_id)
SELECT 1, g.id
FROM goals g
WHERE g.id != 1  -- Don't make root a child of itself
  AND NOT EXISTS (
    SELECT 1
    FROM goal_relations gr
    WHERE gr.child_id = g.id
  )
ON CONFLICT (parent_id, child_id) DO NOTHING;

-- Associate root goal with all users
-- This ensures everyone can see the root goal
INSERT INTO user_goals (user_id, goal_id)
SELECT DISTINCT u.id, 1
FROM users u
WHERE NOT EXISTS (
  SELECT 1
  FROM user_goals ug
  WHERE ug.user_id = u.id AND ug.goal_id = 1
);

COMMIT;
