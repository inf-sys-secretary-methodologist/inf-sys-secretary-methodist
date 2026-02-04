-- ============================================================================
-- STUDENT ANALYTICS VIEWS - SQL views for predictive analytics
-- ============================================================================

-- View: Student attendance statistics
CREATE OR REPLACE VIEW v_student_attendance_stats AS
SELECT
    es.id AS student_id,
    es.full_name AS student_name,
    es.group_name,
    COUNT(DISTINCT ar.id) AS total_records,
    COUNT(DISTINCT CASE WHEN ar.status = 'present' THEN ar.id END) AS present_count,
    COUNT(DISTINCT CASE WHEN ar.status = 'absent' THEN ar.id END) AS absent_count,
    COUNT(DISTINCT CASE WHEN ar.status = 'late' THEN ar.id END) AS late_count,
    COUNT(DISTINCT CASE WHEN ar.status = 'excused' THEN ar.id END) AS excused_count,
    CASE
        WHEN COUNT(DISTINCT ar.id) > 0 THEN
            ROUND(
                (COUNT(DISTINCT CASE WHEN ar.status IN ('present', 'late', 'excused') THEN ar.id END)::DECIMAL /
                 COUNT(DISTINCT ar.id)::DECIMAL) * 100,
                2
            )
        ELSE NULL
    END AS attendance_rate
FROM external_students es
LEFT JOIN attendance_records ar ON ar.student_id = es.id
WHERE es.is_active = true
GROUP BY es.id, es.full_name, es.group_name;

-- View: Student grade statistics
CREATE OR REPLACE VIEW v_student_grade_stats AS
SELECT
    es.id AS student_id,
    es.full_name AS student_name,
    es.group_name,
    COUNT(g.id) AS total_grades,
    ROUND(AVG(g.grade_value), 2) AS grade_average,
    ROUND(
        SUM(g.grade_value * g.weight) / NULLIF(SUM(g.weight), 0),
        2
    ) AS weighted_average,
    MIN(g.grade_value) AS min_grade,
    MAX(g.grade_value) AS max_grade,
    COUNT(CASE WHEN g.grade_value < 50 THEN 1 END) AS failing_grades_count
FROM external_students es
LEFT JOIN grades g ON g.student_id = es.id
WHERE es.is_active = true
GROUP BY es.id, es.full_name, es.group_name;

-- View: Student risk score calculation
CREATE OR REPLACE VIEW v_student_risk_score AS
SELECT
    es.id AS student_id,
    es.full_name AS student_name,
    es.group_name,
    COALESCE(att.attendance_rate, 0) AS attendance_rate,
    COALESCE(grd.grade_average, 0) AS grade_average,
    -- Risk level based on attendance and grades
    CASE
        WHEN COALESCE(att.attendance_rate, 0) < 60 OR COALESCE(grd.grade_average, 0) < 50 THEN 'critical'
        WHEN COALESCE(att.attendance_rate, 0) < 75 OR COALESCE(grd.grade_average, 0) < 65 THEN 'high'
        WHEN COALESCE(att.attendance_rate, 0) < 85 OR COALESCE(grd.grade_average, 0) < 75 THEN 'medium'
        ELSE 'low'
    END AS risk_level,
    -- Risk score (0-100, higher = more at risk)
    CASE
        WHEN att.attendance_rate IS NULL AND grd.grade_average IS NULL THEN 0
        ELSE ROUND(
            100 - (
                (COALESCE(att.attendance_rate, 50) * 0.4) +
                (COALESCE(grd.grade_average, 50) * 0.6)
            ),
            2
        )
    END AS risk_score,
    -- Risk factors as JSON
    jsonb_build_object(
        'attendance', jsonb_build_object(
            'rate', COALESCE(att.attendance_rate, 0),
            'absent_count', COALESCE(att.absent_count, 0),
            'is_risk', COALESCE(att.attendance_rate, 100) < 80
        ),
        'grades', jsonb_build_object(
            'average', COALESCE(grd.grade_average, 0),
            'failing_count', COALESCE(grd.failing_grades_count, 0),
            'is_risk', COALESCE(grd.grade_average, 100) < 70
        )
    ) AS risk_factors,
    att.total_records AS attendance_records_count,
    grd.total_grades AS grades_count
FROM external_students es
LEFT JOIN v_student_attendance_stats att ON att.student_id = es.id
LEFT JOIN v_student_grade_stats grd ON grd.student_id = es.id
WHERE es.is_active = true;

-- View: At-risk students (filtered view for quick access)
CREATE OR REPLACE VIEW v_at_risk_students AS
SELECT *
FROM v_student_risk_score
WHERE risk_level IN ('high', 'critical')
ORDER BY risk_score DESC;

-- View: Group summary statistics
CREATE OR REPLACE VIEW v_group_analytics_summary AS
SELECT
    group_name,
    COUNT(*) AS total_students,
    ROUND(AVG(attendance_rate), 2) AS avg_attendance_rate,
    ROUND(AVG(grade_average), 2) AS avg_grade,
    COUNT(CASE WHEN risk_level = 'critical' THEN 1 END) AS critical_risk_count,
    COUNT(CASE WHEN risk_level = 'high' THEN 1 END) AS high_risk_count,
    COUNT(CASE WHEN risk_level = 'medium' THEN 1 END) AS medium_risk_count,
    COUNT(CASE WHEN risk_level = 'low' THEN 1 END) AS low_risk_count,
    ROUND(
        (COUNT(CASE WHEN risk_level IN ('high', 'critical') THEN 1 END)::DECIMAL /
         NULLIF(COUNT(*), 0)::DECIMAL) * 100,
        2
    ) AS at_risk_percentage
FROM v_student_risk_score
GROUP BY group_name
ORDER BY at_risk_percentage DESC NULLS LAST;

-- View: Monthly attendance trends
CREATE OR REPLACE VIEW v_monthly_attendance_trend AS
SELECT
    DATE_TRUNC('month', ar.lesson_date) AS month,
    COUNT(DISTINCT ar.student_id) AS unique_students,
    COUNT(ar.id) AS total_records,
    COUNT(CASE WHEN ar.status = 'present' THEN 1 END) AS present_count,
    COUNT(CASE WHEN ar.status = 'absent' THEN 1 END) AS absent_count,
    ROUND(
        (COUNT(CASE WHEN ar.status IN ('present', 'late', 'excused') THEN 1 END)::DECIMAL /
         NULLIF(COUNT(ar.id), 0)::DECIMAL) * 100,
        2
    ) AS attendance_rate
FROM attendance_records ar
GROUP BY DATE_TRUNC('month', ar.lesson_date)
ORDER BY month DESC;

-- Comments on views
COMMENT ON VIEW v_student_attendance_stats IS 'Student attendance statistics aggregated by student';
COMMENT ON VIEW v_student_grade_stats IS 'Student grade statistics aggregated by student';
COMMENT ON VIEW v_student_risk_score IS 'Computed risk scores for all students';
COMMENT ON VIEW v_at_risk_students IS 'Filtered view of students with high or critical risk';
COMMENT ON VIEW v_group_analytics_summary IS 'Summary statistics by student group';
COMMENT ON VIEW v_monthly_attendance_trend IS 'Monthly attendance trends for time-series analysis';
