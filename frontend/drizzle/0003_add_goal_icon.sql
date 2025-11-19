-- Migration: Add icon field to goals table

BEGIN;

-- Add icon column to goals with default value
ALTER TABLE goals ADD COLUMN icon varchar(100) DEFAULT 'material-symbols:circle';

COMMIT;
