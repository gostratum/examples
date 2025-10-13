-- Add avatar_url column to users table
ALTER TABLE users ADD COLUMN avatar_url TEXT;

-- Add comment to explain the column
COMMENT ON COLUMN users.avatar_url IS 'URL to user avatar image stored in cloud storage';