-- Create users table for authentication module
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) DEFAULT '',
    role VARCHAR(50) NOT NULL CHECK (role IN ('admin', 'secretary', 'methodist', 'teacher', 'student')),
    status VARCHAR(50) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE INDEX idx_users_email ON users(email);

-- Create index on role for filtering
CREATE INDEX idx_users_role ON users(role);

-- Create index on status
CREATE INDEX idx_users_status ON users(status);

-- Insert default admin user (password: Admin123456!)
-- This is for development only, should be changed in production
INSERT INTO users (email, password, name, role, status)
VALUES ('admin@inf-sys.local', '$2a$14$ZKHqFX3vJT8kZY7ZJy.zfOEzBxD8YmBqGqN1xPJvJ1Y1xYJPqJ5qW', 'System Administrator', 'admin', 'active')
ON CONFLICT (email) DO NOTHING;
