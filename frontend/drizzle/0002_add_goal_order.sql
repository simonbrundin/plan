-- Migration: Add order field to goal_relations for manual sorting

BEGIN;

-- Add order column to goal_relations
ALTER TABLE goal_relations ADD COLUMN "order" integer DEFAULT 0;

-- Set proper ordering for existing relations
-- For each parent, order children by their ID (can be reordered later)
UPDATE goal_relations
SET "order" = (
  SELECT COUNT(*)
  FROM goal_relations gr2
  WHERE gr2.parent_id = goal_relations.parent_id
    AND gr2.child_id <= goal_relations.child_id
) - 1;

COMMIT;
