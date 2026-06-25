-- Migration: Add root goal (ID 1) to all users who don't have it
-- This fixes new users who registered after root goal migration but before the auth fix

BEGIN;

-- Insert root goal associations for users who don't have it
INSERT INTO user_goals (user_id, goal_id)
SELECT DISTINCT u.id, 1
FROM users u
WHERE NOT EXISTS (
    SELECT 1
    FROM user_goals ug
    WHERE ug.user_id = u.id AND ug.goal_id = 1
);

COMMIT;
