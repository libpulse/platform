-- Add unique constraint to prevent duplicate project names for the same user
-- This ensures each user can only have one project with a given name

CREATE UNIQUE INDEX IF NOT EXISTS projects_owner_name_unique
ON projects (owner_user_id, name);

-- Add comment to explain the constraint
COMMENT ON INDEX projects_owner_name_unique IS 'Ensures project names are unique per user';
