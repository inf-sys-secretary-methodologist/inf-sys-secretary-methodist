-- Create sessions table for refresh token management
CREATE TABLE IF NOT EXISTS sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token VARCHAR(500) UNIQUE NOT NULL,
    user_agent VARCHAR(500) DEFAULT '',
    ip_address VARCHAR(45) DEFAULT '',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Create index on refresh_token for faster validation
CREATE INDEX idx_sessions_refresh_token ON sessions(refresh_token);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Create index on composite (user_id, expires_at) for active sessions lookup
CREATE INDEX idx_sessions_user_id_expires_at ON sessions(user_id, expires_at);
