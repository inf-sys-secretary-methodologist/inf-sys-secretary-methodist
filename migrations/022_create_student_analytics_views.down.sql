-- Rollback: Remove student analytics views

DROP VIEW IF EXISTS v_monthly_attendance_trend;
DROP VIEW IF EXISTS v_group_analytics_summary;
DROP VIEW IF EXISTS v_at_risk_students;
DROP VIEW IF EXISTS v_student_risk_score;
DROP VIEW IF EXISTS v_student_grade_stats;
DROP VIEW IF EXISTS v_student_attendance_stats;
