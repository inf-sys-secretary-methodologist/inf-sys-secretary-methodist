-- Migration: Create Integration Module Schema
-- Description: Tables for 1C system integration (employees, students, sync logs, conflicts)

-- Sync status enum
CREATE TYPE sync_status AS ENUM ('pending', 'in_progress', 'completed', 'failed', 'cancelled');

-- Sync direction enum
CREATE TYPE sync_direction AS ENUM ('import', 'export', 'both');

-- Sync entity type enum
CREATE TYPE sync_entity_type AS ENUM ('employee', 'student', 'finance');

-- Conflict resolution enum
CREATE TYPE conflict_resolution AS ENUM ('pending', 'use_local', 'use_external', 'merge', 'skip');

-- Sync logs table
CREATE TABLE IF NOT EXISTS sync_logs (
    id BIGSERIAL PRIMARY KEY,
    entity_type sync_entity_type NOT NULL,
    direction sync_direction NOT NULL,
    status sync_status NOT NULL DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    total_records INTEGER NOT NULL DEFAULT 0,
    processed_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    conflict_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- External employees table (data from 1C)
CREATE TABLE IF NOT EXISTS external_employees (
    id BIGSERIAL PRIMARY KEY,
    external_id VARCHAR(50) NOT NULL UNIQUE,  -- 1C GUID (Ref_Key)
    code VARCHAR(50) NOT NULL,                 -- 1C Code
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    position VARCHAR(255),
    department VARCHAR(255),
    employment_date DATE,
    dismissal_date DATE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    local_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    last_sync_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    external_data_hash VARCHAR(64),           -- SHA256 hash for change detection
    raw_data JSONB,                            -- Original JSON from 1C
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- External students table (data from 1C)
CREATE TABLE IF NOT EXISTS external_students (
    id BIGSERIAL PRIMARY KEY,
    external_id VARCHAR(50) NOT NULL UNIQUE,  -- 1C GUID (Ref_Key)
    code VARCHAR(50) NOT NULL,                 -- 1C Code (зачетка)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(50),
    group_name VARCHAR(100),
    faculty VARCHAR(255),
    specialty VARCHAR(255),
    course INTEGER,
    study_form VARCHAR(100),
    enrollment_date DATE,
    expulsion_date DATE,
    graduation_date DATE,
    status VARCHAR(50) NOT NULL DEFAULT 'enrolled',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    local_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    last_sync_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    external_data_hash VARCHAR(64),
    raw_data JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Sync conflicts table
CREATE TABLE IF NOT EXISTS sync_conflicts (
    id BIGSERIAL PRIMARY KEY,
    sync_log_id BIGINT NOT NULL REFERENCES sync_logs(id) ON DELETE CASCADE,
    entity_type sync_entity_type NOT NULL,
    entity_id VARCHAR(50) NOT NULL,            -- External ID or local ID
    local_data JSONB,                          -- JSON of local record
    external_data JSONB,                       -- JSON of external record
    conflict_type VARCHAR(50) NOT NULL,        -- update, delete, create
    conflict_fields TEXT[],                    -- Fields with conflicts
    resolution conflict_resolution NOT NULL DEFAULT 'pending',
    resolved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolved_data JSONB,                       -- Final merged data
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for sync_logs
CREATE INDEX idx_sync_logs_entity_type ON sync_logs(entity_type);
CREATE INDEX idx_sync_logs_status ON sync_logs(status);
CREATE INDEX idx_sync_logs_started_at ON sync_logs(started_at DESC);
CREATE INDEX idx_sync_logs_entity_status ON sync_logs(entity_type, status);

-- Indexes for external_employees
CREATE INDEX idx_external_employees_code ON external_employees(code);
CREATE INDEX idx_external_employees_email ON external_employees(email) WHERE email IS NOT NULL;
CREATE INDEX idx_external_employees_local_user_id ON external_employees(local_user_id) WHERE local_user_id IS NOT NULL;
CREATE INDEX idx_external_employees_is_active ON external_employees(is_active);
CREATE INDEX idx_external_employees_department ON external_employees(department) WHERE department IS NOT NULL;
CREATE INDEX idx_external_employees_last_sync ON external_employees(last_sync_at);

-- Full-text search for employees
CREATE INDEX idx_external_employees_search ON external_employees
    USING GIN (to_tsvector('russian', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(middle_name, '')));

-- Indexes for external_students
CREATE INDEX idx_external_students_code ON external_students(code);
CREATE INDEX idx_external_students_email ON external_students(email) WHERE email IS NOT NULL;
CREATE INDEX idx_external_students_local_user_id ON external_students(local_user_id) WHERE local_user_id IS NOT NULL;
CREATE INDEX idx_external_students_is_active ON external_students(is_active);
CREATE INDEX idx_external_students_group ON external_students(group_name) WHERE group_name IS NOT NULL;
CREATE INDEX idx_external_students_faculty ON external_students(faculty) WHERE faculty IS NOT NULL;
CREATE INDEX idx_external_students_course ON external_students(course) WHERE course IS NOT NULL;
CREATE INDEX idx_external_students_status ON external_students(status);
CREATE INDEX idx_external_students_last_sync ON external_students(last_sync_at);

-- Full-text search for students
CREATE INDEX idx_external_students_search ON external_students
    USING GIN (to_tsvector('russian', coalesce(first_name, '') || ' ' || coalesce(last_name, '') || ' ' || coalesce(middle_name, '')));

-- Indexes for sync_conflicts
CREATE INDEX idx_sync_conflicts_sync_log_id ON sync_conflicts(sync_log_id);
CREATE INDEX idx_sync_conflicts_entity_type ON sync_conflicts(entity_type);
CREATE INDEX idx_sync_conflicts_resolution ON sync_conflicts(resolution);
CREATE INDEX idx_sync_conflicts_pending ON sync_conflicts(resolution) WHERE resolution = 'pending';
CREATE INDEX idx_sync_conflicts_created_at ON sync_conflicts(created_at DESC);

-- Updated_at trigger function (if not exists)
CREATE OR REPLACE FUNCTION update_integration_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Updated_at triggers
CREATE TRIGGER update_sync_logs_updated_at
    BEFORE UPDATE ON sync_logs
    FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at_column();

CREATE TRIGGER update_external_employees_updated_at
    BEFORE UPDATE ON external_employees
    FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at_column();

CREATE TRIGGER update_external_students_updated_at
    BEFORE UPDATE ON external_students
    FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at_column();

CREATE TRIGGER update_sync_conflicts_updated_at
    BEFORE UPDATE ON sync_conflicts
    FOR EACH ROW EXECUTE FUNCTION update_integration_updated_at_column();

-- Comments
COMMENT ON TABLE sync_logs IS 'Logs of synchronization operations with 1C system';
COMMENT ON TABLE external_employees IS 'Employee records synchronized from 1C system';
COMMENT ON TABLE external_students IS 'Student records synchronized from 1C system';
COMMENT ON TABLE sync_conflicts IS 'Data conflicts detected during synchronization';

COMMENT ON COLUMN external_employees.external_id IS '1C GUID (Ref_Key) - unique identifier in 1C system';
COMMENT ON COLUMN external_employees.code IS '1C Code - employee code in 1C system';
COMMENT ON COLUMN external_employees.external_data_hash IS 'SHA256 hash of external data for change detection';
COMMENT ON COLUMN external_employees.local_user_id IS 'Link to local user account if exists';

COMMENT ON COLUMN external_students.external_id IS '1C GUID (Ref_Key) - unique identifier in 1C system';
COMMENT ON COLUMN external_students.code IS '1C Code (зачетка) - student ID in 1C system';
COMMENT ON COLUMN external_students.status IS 'Student status: enrolled, graduated, expelled, academic_leave';
