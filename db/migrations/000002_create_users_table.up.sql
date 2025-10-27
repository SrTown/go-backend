-- Create users table (CockroachDB optimized)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email STRING UNIQUE NOT NULL, 
    name STRING NOT NULL,
    password STRING NOT NULL,
    user_type STRING NOT NULL,
    status BOOLEAN,
    created_at TIMESTAMP DEFAULT current_timestamp(),
    updated_at TIMESTAMP DEFAULT current_timestamp()
);

-- Index for faster queries by email (most common lookup)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Index for filtering by user type
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);

-- Index for date-based queries (storing order in CockroachDB)
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC) STORING (email, name);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = current_timestamp();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();