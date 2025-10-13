-- Remove avatar_url column from users table
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;