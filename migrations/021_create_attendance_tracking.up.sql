-- ============================================================================
-- ATTENDANCE TRACKING - Tables for student attendance and grades
-- ============================================================================

-- Lessons/Classes table for tracking what can be attended
CREATE TABLE IF NOT EXISTS lessons (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    teacher_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    group_name VARCHAR(100), -- Student group (e.g., "ИС-21")
    lesson_type VARCHAR(50) DEFAULT 'lecture' CHECK (lesson_type IN ('lecture', 'practice', 'lab', 'seminar', 'exam')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lessons_teacher_id ON lessons(teacher_id);
CREATE INDEX idx_lessons_group_name ON lessons(group_name);
CREATE INDEX idx_lessons_subject ON lessons(subject);

-- Attendance records
CREATE TABLE IF NOT EXISTS attendance_records (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL, -- References external_students.id
    lesson_id BIGINT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    lesson_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'absent' CHECK (status IN ('present', 'absent', 'late', 'excused')),
    marked_by BIGINT REFERENCES users(id) ON DELETE SET NULL, -- Who marked attendance
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(student_id, lesson_id, lesson_date)
);

CREATE INDEX idx_attendance_records_student_id ON attendance_records(student_id);
CREATE INDEX idx_attendance_records_lesson_id ON attendance_records(lesson_id);
CREATE INDEX idx_attendance_records_lesson_date ON attendance_records(lesson_date);
CREATE INDEX idx_attendance_records_status ON attendance_records(status);
CREATE INDEX idx_attendance_records_student_date ON attendance_records(student_id, lesson_date);

-- Grades table
CREATE TABLE IF NOT EXISTS grades (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL, -- References external_students.id
    subject VARCHAR(255) NOT NULL,
    grade_type VARCHAR(50) NOT NULL DEFAULT 'current' CHECK (grade_type IN ('current', 'midterm', 'final', 'test', 'homework')),
    grade_value DECIMAL(5,2) NOT NULL CHECK (grade_value >= 0 AND grade_value <= 100),
    max_value DECIMAL(5,2) NOT NULL DEFAULT 100,
    weight DECIMAL(3,2) NOT NULL DEFAULT 1.0, -- Weight for weighted average
    graded_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    grade_date DATE NOT NULL DEFAULT CURRENT_DATE,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_grades_student_id ON grades(student_id);
CREATE INDEX idx_grades_subject ON grades(subject);
CREATE INDEX idx_grades_grade_date ON grades(grade_date);
CREATE INDEX idx_grades_grade_type ON grades(grade_type);
CREATE INDEX idx_grades_student_subject ON grades(student_id, subject);

-- Student risk assessments (cached/computed data)
CREATE TABLE IF NOT EXISTS student_risk_assessments (
    id BIGSERIAL PRIMARY KEY,
    student_id BIGINT NOT NULL UNIQUE, -- References external_students.id
    attendance_rate DECIMAL(5,2), -- 0-100%
    grade_average DECIMAL(5,2), -- 0-100
    risk_level VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (risk_level IN ('low', 'medium', 'high', 'critical', 'unknown')),
    risk_score DECIMAL(5,2), -- 0-100, higher = more at risk
    risk_factors JSONB, -- Array of risk factor details
    last_calculated_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_student_risk_assessments_risk_level ON student_risk_assessments(risk_level);
CREATE INDEX idx_student_risk_assessments_risk_score ON student_risk_assessments(risk_score DESC);

-- Comments on tables
COMMENT ON TABLE lessons IS 'Lessons/classes that students can attend';
COMMENT ON TABLE attendance_records IS 'Student attendance records for each lesson';
COMMENT ON TABLE grades IS 'Student grades for various assessments';
COMMENT ON TABLE student_risk_assessments IS 'Cached risk assessment data for students';

-- Trigger to update updated_at
CREATE OR REPLACE FUNCTION update_attendance_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_lessons_updated_at
    BEFORE UPDATE ON lessons
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();

CREATE TRIGGER tr_attendance_records_updated_at
    BEFORE UPDATE ON attendance_records
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();

CREATE TRIGGER tr_grades_updated_at
    BEFORE UPDATE ON grades
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();

CREATE TRIGGER tr_student_risk_assessments_updated_at
    BEFORE UPDATE ON student_risk_assessments
    FOR EACH ROW
    EXECUTE FUNCTION update_attendance_updated_at();
