-- Rollback: Remove attendance tracking tables

DROP TRIGGER IF EXISTS tr_student_risk_assessments_updated_at ON student_risk_assessments;
DROP TRIGGER IF EXISTS tr_grades_updated_at ON grades;
DROP TRIGGER IF EXISTS tr_attendance_records_updated_at ON attendance_records;
DROP TRIGGER IF EXISTS tr_lessons_updated_at ON lessons;

DROP FUNCTION IF EXISTS update_attendance_updated_at();

DROP TABLE IF EXISTS student_risk_assessments;
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS attendance_records;
DROP TABLE IF EXISTS lessons;
