-- Risk weight configuration table (admin-configurable)
CREATE TABLE IF NOT EXISTS risk_weight_config (
    id SERIAL PRIMARY KEY,
    attendance_weight NUMERIC(4,2) NOT NULL DEFAULT 0.35,
    grade_weight NUMERIC(4,2) NOT NULL DEFAULT 0.30,
    submission_weight NUMERIC(4,2) NOT NULL DEFAULT 0.20,
    inactivity_weight NUMERIC(4,2) NOT NULL DEFAULT 0.15,
    high_risk_threshold NUMERIC(4,1) NOT NULL DEFAULT 70.0,
    critical_risk_threshold NUMERIC(4,1) NOT NULL DEFAULT 85.0,
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_weights_sum CHECK (
        ABS(attendance_weight + grade_weight + submission_weight + inactivity_weight - 1.0) < 0.01
    ),
    CONSTRAINT chk_weights_positive CHECK (
        attendance_weight >= 0 AND grade_weight >= 0 AND submission_weight >= 0 AND inactivity_weight >= 0
    )
);

-- Insert default configuration
INSERT INTO risk_weight_config (attendance_weight, grade_weight, submission_weight, inactivity_weight)
VALUES (0.35, 0.30, 0.20, 0.15)
ON CONFLICT DO NOTHING;

-- Student risk history table (daily snapshots)
CREATE TABLE IF NOT EXISTS student_risk_history (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    risk_score NUMERIC(5,2) NOT NULL,
    risk_level VARCHAR(20) NOT NULL,
    attendance_rate NUMERIC(5,2),
    grade_average NUMERIC(5,2),
    submission_rate NUMERIC(5,2),
    risk_factors JSONB,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_risk_history_student_date ON student_risk_history (student_id, calculated_at DESC);
CREATE INDEX idx_risk_history_calculated_at ON student_risk_history (calculated_at DESC);
CREATE INDEX idx_risk_history_risk_level ON student_risk_history (risk_level);
