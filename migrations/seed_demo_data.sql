-- ============================================================================
-- DEMO SEED DATA — для демонстрации системы научному руководителю
-- Пароль у всех пользователей: 12345678
-- Запуск: psql -d <dbname> -f migrations/seed_demo_data.sql
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ПОЛЬЗОВАТЕЛИ (5 ролей, 10 пользователей)
-- ============================================================================

-- Пароль 123456 bcrypt hash
-- $2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i

INSERT INTO users (email, password, name, role, status, created_at, updated_at) VALUES
-- Администратор
('admin@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Петров Алексей Иванович', 'system_admin', 'active', NOW() - INTERVAL '180 days', NOW()),
-- Методисты
('methodist@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Козлова Елена Сергеевна', 'methodist', 'active', NOW() - INTERVAL '150 days', NOW()),
-- Секретари
('secretary@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Новикова Ольга Дмитриевна', 'academic_secretary', 'active', NOW() - INTERVAL '150 days', NOW()),
-- Преподаватели
('ivanov@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Иванов Дмитрий Николаевич', 'teacher', 'active', NOW() - INTERVAL '120 days', NOW()),
('sidorova@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Сидорова Анна Петровна', 'teacher', 'active', NOW() - INTERVAL '120 days', NOW()),
('kuznetsov@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Кузнецов Михаил Александрович', 'teacher', 'active', NOW() - INTERVAL '90 days', NOW()),
-- Студенты
('student1@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Смирнов Артём Владимирович', 'student', 'active', NOW() - INTERVAL '60 days', NOW()),
('student2@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Волкова Мария Андреевна', 'student', 'active', NOW() - INTERVAL '60 days', NOW()),
('student3@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Морозов Никита Сергеевич', 'student', 'active', NOW() - INTERVAL '45 days', NOW()),
('student4@university.ru', '$2a$10$m7NB00vZiVZ.GhnqHjFvGupA9/ledZXNVl.EPHkVdXkf643pN3M0i', 'Лебедева Дарья Игоревна', 'student', 'active', NOW() - INTERVAL '45 days', NOW())
ON CONFLICT (email) DO NOTHING;

-- ============================================================================
-- 2. ОРГАНИЗАЦИОННАЯ СТРУКТУРА (org_departments, positions, user_profiles)
-- ============================================================================

INSERT INTO org_departments (name, code, description, is_active) VALUES
('Ректорат', 'RECT', 'Руководство университета', true),
('Учебно-методическое управление', 'UMU', 'Методическое обеспечение учебного процесса', true),
('Учебный отдел', 'UO', 'Организация учебного процесса', true),
('Кафедра информатики', 'KINF', 'Кафедра информатики и вычислительной техники', true),
('Кафедра экономики', 'KECON', 'Кафедра экономики и управления', true),
('Кафедра математики', 'KMATH', 'Кафедра высшей математики', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO positions (name, code, description, level, is_active) VALUES
('Системный администратор', 'SYSADMIN', 'Администратор информационных систем', 90, true),
('Методист', 'METHODIST', 'Методист учебно-методического управления', 70, true),
('Академический секретарь', 'ACADSEC', 'Секретарь учёного совета / деканата', 70, true),
('Доцент', 'DOCENT', 'Преподаватель (доцент)', 60, true),
('Старший преподаватель', 'SENIOR_LECTURER', 'Старший преподаватель', 50, true),
('Студент', 'STUDENT', 'Обучающийся', 10, true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO user_profiles (user_id, department_id, position_id, phone, bio)
SELECT u.id, d.id, p.id, phone, bio FROM (VALUES
  ('admin@university.ru', 'RECT', 'SYSADMIN', '+7 (999) 100-00-01', 'Администратор ИС университета'),
  ('methodist@university.ru', 'UMU', 'METHODIST', '+7 (999) 200-00-01', 'Методист УМУ, стаж 8 лет'),
  ('secretary@university.ru', 'UO', 'ACADSEC', '+7 (999) 300-00-01', 'Академический секретарь факультета'),
  ('ivanov@university.ru', 'KINF', 'DOCENT', '+7 (999) 400-00-01', 'Доцент кафедры информатики, к.т.н.'),
  ('sidorova@university.ru', 'KECON', 'SENIOR_LECTURER', '+7 (999) 400-00-02', 'Преподаватель кафедры экономики'),
  ('kuznetsov@university.ru', 'KMATH', 'DOCENT', '+7 (999) 400-00-03', 'Доцент кафедры математики, к.ф.-м.н.')
) AS v(email, dept_code, pos_code, phone, bio)
JOIN users u ON u.email = v.email
JOIN org_departments d ON d.code = v.dept_code
JOIN positions p ON p.code = v.pos_code
ON CONFLICT (user_id) DO NOTHING;

-- ============================================================================
-- 3. АКАДЕМИЧЕСКАЯ СТРУКТУРА (faculties, departments, specialties, groups)
-- ============================================================================

INSERT INTO faculties (name, short_name, code) VALUES
('Факультет информационных технологий', 'ФИТ', 'FIT'),
('Факультет экономики и управления', 'ФЭУ', 'FEU')
ON CONFLICT (name) DO NOTHING;

INSERT INTO departments (faculty_id, name, short_name, code)
SELECT f.id, d.name, d.short_name, d.code FROM (VALUES
  ('FIT', 'Кафедра информатики и ВТ', 'КИиВТ', 'KIVT'),
  ('FIT', 'Кафедра программной инженерии', 'КПИ', 'KPI'),
  ('FEU', 'Кафедра экономики', 'КЭ', 'KE'),
  ('FEU', 'Кафедра менеджмента', 'КМ', 'KM')
) AS d(faculty_code, name, short_name, code)
JOIN faculties f ON f.code = d.faculty_code
ON CONFLICT (faculty_id, code) DO NOTHING;

INSERT INTO specialties (faculty_id, name, code, qualification, duration_years)
SELECT f.id, s.name, s.code, s.qualification, s.duration FROM (VALUES
  ('FIT', 'Информатика и вычислительная техника', '09.03.01', 'Бакалавр', 4),
  ('FIT', 'Программная инженерия', '09.03.04', 'Бакалавр', 4),
  ('FEU', 'Экономика', '38.03.01', 'Бакалавр', 4),
  ('FEU', 'Менеджмент', '38.03.02', 'Бакалавр', 4)
) AS s(faculty_code, name, code, qualification, duration)
JOIN faculties f ON f.code = s.faculty_code;

-- Группы студентов
INSERT INTO student_groups (specialty_id, name, course, capacity)
SELECT sp.id, g.name, g.course, g.capacity FROM (VALUES
  ('09.03.01', 'ИВТ-201', 2, 25),
  ('09.03.01', 'ИВТ-202', 2, 25),
  ('09.03.04', 'ПИ-301', 3, 20),
  ('38.03.01', 'ЭК-201', 2, 30),
  ('38.03.02', 'МН-101', 1, 28)
) AS g(spec_code, name, course, capacity)
JOIN specialties sp ON sp.code = g.spec_code;

-- Привязать кураторов к группам
UPDATE student_groups SET curator_id = (SELECT id FROM users WHERE email = 'ivanov@university.ru')
WHERE name IN ('ИВТ-201', 'ИВТ-202');
UPDATE student_groups SET curator_id = (SELECT id FROM users WHERE email = 'sidorova@university.ru')
WHERE name = 'ЭК-201';

-- ============================================================================
-- 4. УЧЕБНЫЙ ГОД И СЕМЕСТР
-- ============================================================================

INSERT INTO academic_years (name, start_date, end_date, is_active) VALUES
('2025/2026', '2025-09-01', '2026-06-30', true),
('2024/2025', '2024-09-01', '2025-06-30', false)
ON CONFLICT (name) DO NOTHING;

INSERT INTO semesters (academic_year_id, name, number, start_date, end_date, is_active)
SELECT ay.id, s.name, s.number, s.start_date::date, s.end_date::date, s.is_active FROM (VALUES
  ('2025/2026', 'Осенний семестр 2025/2026', 1, '2025-09-01', '2025-12-31', false),
  ('2025/2026', 'Весенний семестр 2025/2026', 2, '2026-02-01', '2026-06-30', true)
) AS s(year_name, name, number, start_date, end_date, is_active)
JOIN academic_years ay ON ay.name = s.year_name
ON CONFLICT (academic_year_id, number) DO NOTHING;

-- ============================================================================
-- 5. ДИСЦИПЛИНЫ
-- ============================================================================

INSERT INTO disciplines (name, code, department_id, credits, hours_total, hours_lectures, hours_practice, hours_labs, control_type) VALUES
('Информатика', 'INF-101', (SELECT id FROM departments WHERE code = 'KIVT'), 5, 180, 54, 36, 36, 'exam'),
('Базы данных', 'DB-201', (SELECT id FROM departments WHERE code = 'KIVT'), 4, 144, 36, 36, 36, 'exam'),
('Программная инженерия', 'SE-301', (SELECT id FROM departments WHERE code = 'KPI'), 5, 180, 54, 54, 0, 'exam'),
('Математический анализ', 'MA-101', (SELECT id FROM departments WHERE code = 'KIVT'), 6, 216, 72, 72, 0, 'exam'),
('Микроэкономика', 'MICRO-201', (SELECT id FROM departments WHERE code = 'KE'), 4, 144, 54, 36, 0, 'exam'),
('Менеджмент', 'MGT-101', (SELECT id FROM departments WHERE code = 'KM'), 3, 108, 36, 36, 0, 'test'),
('Английский язык', 'ENG-101', NULL, 2, 72, 0, 72, 0, 'test'),
('Физическая культура', 'PE-101', NULL, 0, 72, 0, 0, 0, 'test');

-- ============================================================================
-- 6. АУДИТОРИИ
-- ============================================================================

INSERT INTO classrooms (building, number, name, capacity, type, equipment, is_available) VALUES
('Корпус А', '101', 'Большая лекционная аудитория', 200, 'lecture', '{"projector": true, "microphone": true, "screen": true}', true),
('Корпус А', '205', 'Аудитория 205', 40, 'practice', '{"projector": true, "whiteboard": true}', true),
('Корпус А', '301', 'Компьютерный класс №1', 25, 'computer', '{"computers": 25, "projector": true, "printer": true}', true),
('Корпус А', '302', 'Компьютерный класс №2', 25, 'computer', '{"computers": 25, "projector": true}', true),
('Корпус Б', '101', 'Лабораторная №1', 20, 'lab', '{"projector": true, "lab_equipment": true}', true),
('Корпус Б', '203', 'Аудитория 203', 35, 'practice', '{"projector": true, "whiteboard": true}', true),
('Корпус Б', '305', 'Конференц-зал', 60, 'conference', '{"projector": true, "microphone": true, "video_conference": true}', true),
('Спорткомплекс', '1', 'Спортивный зал', 100, 'sports', '{}', true)
ON CONFLICT (building, number) DO NOTHING;

-- ============================================================================
-- 7. РАСПИСАНИЕ ЗАНЯТИЙ (весенний семестр 2025/2026)
-- ============================================================================

INSERT INTO schedule_lessons (semester_id, discipline_id, lesson_type_id, teacher_id, group_id, classroom_id, day_of_week, time_start, time_end, week_type, date_start, date_end)
SELECT
  sem.id, disc.id, lt.id, teacher.id, grp.id, room.id,
  v.day_of_week, v.time_start::time, v.time_end::time, v.week_type,
  '2026-02-01'::date, '2026-06-30'::date
FROM (VALUES
  -- Понедельник
  ('Базы данных', 'Лекция', 'ivanov@university.ru', 'ИВТ-201', 'Корпус А', '101', 1, '09:00', '10:30', 'all'),
  ('Математический анализ', 'Лекция', 'kuznetsov@university.ru', 'ИВТ-201', 'Корпус А', '205', 1, '10:45', '12:15', 'all'),
  ('Микроэкономика', 'Лекция', 'sidorova@university.ru', 'ЭК-201', 'Корпус Б', '203', 1, '09:00', '10:30', 'all'),
  ('Менеджмент', 'Практическое занятие', 'sidorova@university.ru', 'ЭК-201', 'Корпус Б', '203', 1, '10:45', '12:15', 'all'),
  -- Вторник
  ('Базы данных', 'Лабораторная работа', 'ivanov@university.ru', 'ИВТ-201', 'Корпус А', '301', 2, '09:00', '10:30', 'odd'),
  ('Базы данных', 'Практическое занятие', 'ivanov@university.ru', 'ИВТ-201', 'Корпус А', '205', 2, '09:00', '10:30', 'even'),
  ('Программная инженерия', 'Лекция', 'ivanov@university.ru', 'ПИ-301', 'Корпус А', '101', 2, '10:45', '12:15', 'all'),
  ('Английский язык', 'Практическое занятие', 'kuznetsov@university.ru', 'ИВТ-201', 'Корпус А', '205', 2, '13:00', '14:30', 'all'),
  -- Среда
  ('Информатика', 'Лекция', 'ivanov@university.ru', 'ИВТ-201', 'Корпус А', '101', 3, '09:00', '10:30', 'all'),
  ('Информатика', 'Лабораторная работа', 'ivanov@university.ru', 'ИВТ-201', 'Корпус А', '302', 3, '10:45', '12:15', 'all'),
  ('Микроэкономика', 'Семинар', 'sidorova@university.ru', 'ЭК-201', 'Корпус Б', '203', 3, '09:00', '10:30', 'all'),
  -- Четверг
  ('Математический анализ', 'Практическое занятие', 'kuznetsov@university.ru', 'ИВТ-201', 'Корпус А', '205', 4, '09:00', '10:30', 'all'),
  ('Программная инженерия', 'Практическое занятие', 'ivanov@university.ru', 'ПИ-301', 'Корпус А', '301', 4, '10:45', '12:15', 'all'),
  ('Менеджмент', 'Лекция', 'sidorova@university.ru', 'МН-101', 'Корпус Б', '203', 4, '09:00', '10:30', 'all'),
  -- Пятница
  ('Физическая культура', 'Практическое занятие', 'kuznetsov@university.ru', 'ИВТ-201', 'Спорткомплекс', '1', 5, '09:00', '10:30', 'all'),
  ('Базы данных', 'Семинар', 'ivanov@university.ru', 'ИВТ-202', 'Корпус А', '205', 5, '10:45', '12:15', 'all'),
  -- Суббота
  ('Английский язык', 'Практическое занятие', 'kuznetsov@university.ru', 'ЭК-201', 'Корпус Б', '203', 6, '09:00', '10:30', 'even')
) AS v(disc_name, lt_name, teacher_email, group_name, building, room_number, day_of_week, time_start, time_end, week_type)
JOIN semesters sem ON sem.is_active = true
JOIN disciplines disc ON disc.name = v.disc_name
JOIN lesson_types lt ON lt.name = v.lt_name
JOIN users teacher ON teacher.email = v.teacher_email
JOIN student_groups grp ON grp.name = v.group_name
JOIN classrooms room ON room.building = v.building AND room.number = v.room_number;

-- Замена: одна пара отменена
INSERT INTO schedule_changes (lesson_id, change_type, original_date, reason, created_by, created_at)
SELECT sl.id, 'cancelled', '2026-04-21', 'Преподаватель на конференции',
  (SELECT id FROM users WHERE email = 'secretary@university.ru'), NOW()
FROM schedule_lessons sl
JOIN disciplines d ON d.id = sl.discipline_id
WHERE d.name = 'Базы данных' AND sl.day_of_week = 1
LIMIT 1;

-- ============================================================================
-- 8. ДОКУМЕНТЫ
-- ============================================================================

-- Типы документов (seed)
INSERT INTO document_types (name, code, description, requires_approval) VALUES
('Служебная записка', 'memo', 'Внутренний документ', false),
('Приказ (основная деятельность)', 'order_main', 'Приказ по основной деятельности', true),
('Протокол', 'protocol', 'Протокол заседания', false),
('Рабочая программа дисциплины', 'syllabus', 'РПД по ФГОС', true),
('Положение', 'regulation', 'Нормативный документ', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO document_categories (name, description) VALUES
('Учебный процесс', 'Документы по организации учебного процесса'),
('Кадровые', 'Документы по личному составу'),
('Нормативные', 'Положения, регламенты, стандарты')
ON CONFLICT DO NOTHING;

INSERT INTO documents (document_type_id, category_id, title, content, author_id, status, importance, is_public, created_at, updated_at)
SELECT dt.id, dc.id, v.title, v.content, u.id, v.status, v.importance, v.is_public, NOW() - v.days_ago * INTERVAL '1 day', NOW()
FROM (VALUES
  ('memo', 'Учебный процесс', 'Об изменении расписания весеннего семестра', 'В связи с ремонтом корпуса Б, часть занятий переносится в корпус А с 15 марта.', 'secretary@university.ru', 'approved', 'high', true, 30),
  ('order_main', 'Учебный процесс', 'Об утверждении расписания экзаменов', 'Утвердить расписание экзаменационной сессии весеннего семестра 2025/2026 уч. года.', 'admin@university.ru', 'approved', 'urgent', true, 14),
  ('protocol', 'Нормативные', 'Протокол заседания кафедры информатики №7', 'Рассмотрены вопросы: 1) Утверждение тем ВКР 2) Распределение нагрузки на следующий год.', 'ivanov@university.ru', 'registered', 'normal', false, 7),
  ('syllabus', 'Учебный процесс', 'РПД: Базы данных (09.03.01)', 'Рабочая программа дисциплины "Базы данных" для направления 09.03.01 ИВТ.', 'methodist@university.ru', 'approved', 'normal', true, 60),
  ('syllabus', 'Учебный процесс', 'РПД: Программная инженерия (09.03.04)', 'Рабочая программа дисциплины "Программная инженерия" для направления 09.03.04.', 'methodist@university.ru', 'draft', 'normal', false, 3),
  ('regulation', 'Нормативные', 'Положение о текущем контроле успеваемости', 'Устанавливает порядок проведения текущего контроля и промежуточной аттестации.', 'methodist@university.ru', 'approved', 'high', true, 90),
  ('memo', 'Учебный процесс', 'О проведении дня открытых дверей', 'Прошу обеспечить участие студентов 2 курса в Дне открытых дверей 15 мая 2026 г.', 'secretary@university.ru', 'routing', 'normal', false, 2)
) AS v(type_code, cat_name, title, content, author_email, status, importance, is_public, days_ago)
JOIN document_types dt ON dt.code = v.type_code
JOIN document_categories dc ON dc.name = v.cat_name
JOIN users u ON u.email = v.author_email;

-- Теги для документов
INSERT INTO document_tags (name, color) VALUES
('Расписание', '#3B82F6'),
('ФГОС', '#10B981'),
('Экзамены', '#EF4444'),
('Методика', '#8B5CF6'),
('Срочно', '#F59E0B')
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- 9. ЗАДАНИЯ (tasks)
-- ============================================================================

INSERT INTO projects (name, description, owner_id, status, start_date, end_date)
SELECT 'Подготовка к аккредитации', 'Сбор и подготовка документов для аккредитации образовательных программ', u.id, 'active', '2026-03-01', '2026-06-30'
FROM users u WHERE u.email = 'methodist@university.ru';

INSERT INTO tasks (project_id, title, description, author_id, assignee_id, status, priority, due_date, progress, tags, created_at, updated_at)
SELECT
  p.id, v.title, v.description, author.id, assignee.id, v.status, v.priority,
  (NOW() + v.due_days * INTERVAL '1 day')::timestamp, v.progress, v.tags,
  NOW() - v.created_ago * INTERVAL '1 day', NOW()
FROM (VALUES
  ('Обновить РПД по информатике', 'Актуализировать рабочую программу дисциплины в соответствии с новым ФГОС', 'methodist@university.ru', 'ivanov@university.ru', 'in_progress', 'high', 14, 60, ARRAY['ФГОС', 'РПД'], 20),
  ('Подготовить отчёт по посещаемости', 'Сформировать отчёт о посещаемости за март для деканата', 'secretary@university.ru', 'secretary@university.ru', 'completed', 'normal', -5, 100, ARRAY['отчёт', 'посещаемость'], 15),
  ('Составить расписание экзаменов', 'Подготовить проект расписания экзаменационной сессии', 'admin@university.ru', 'secretary@university.ru', 'review', 'urgent', 7, 80, ARRAY['экзамены', 'расписание'], 10),
  ('Проверить лабораторные работы', 'Проверить лабораторные работы студентов ИВТ-201 по курсу БД', 'ivanov@university.ru', 'ivanov@university.ru', 'in_progress', 'normal', 5, 40, ARRAY['лабораторные', 'БД'], 7),
  ('Загрузить презентации лекций', 'Разместить материалы лекций по микроэкономике в системе', 'sidorova@university.ru', 'sidorova@university.ru', 'new', 'low', 21, 0, ARRAY['материалы'], 2),
  ('Провести анализ успеваемости', 'Подготовить аналитику по успеваемости студентов 2 курса ФИТ', 'methodist@university.ru', 'methodist@university.ru', 'assigned', 'high', 10, 0, ARRAY['аналитика', 'успеваемость'], 5)
) AS v(title, description, author_email, assignee_email, status, priority, due_days, progress, tags, created_ago)
JOIN projects p ON p.name = 'Подготовка к аккредитации'
JOIN users author ON author.email = v.author_email
JOIN users assignee ON assignee.email = v.assignee_email;

-- ============================================================================
-- 10. ОБЪЯВЛЕНИЯ
-- ============================================================================

INSERT INTO announcements (title, content, summary, author_id, status, priority, target_audience, is_pinned, view_count, tags, created_at, updated_at, publish_at)
SELECT v.title, v.content, v.summary, u.id, v.status, v.priority, v.audience, v.pinned, v.views, v.tags,
  NOW() - v.days_ago * INTERVAL '1 day', NOW(), NOW() - v.days_ago * INTERVAL '1 day'
FROM (VALUES
  ('Изменения в расписании с 15 марта', 'Уважаемые студенты и преподаватели! В связи с плановым ремонтом корпуса Б, с 15 марта часть занятий переносится в корпус А. Обновлённое расписание доступно в разделе "Расписание".', 'Перенос занятий из корпуса Б в корпус А', 'secretary@university.ru', 'published', 'high', 'all', true, 156, ARRAY['расписание', 'перенос'], 20),
  ('День открытых дверей — 15 мая', 'Приглашаем абитуриентов и их родителей на День открытых дверей! Начало в 10:00 в конференц-зале корпуса Б. Программа: презентация факультетов, экскурсия, встреча с преподавателями.', 'День открытых дверей 15 мая в 10:00', 'admin@university.ru', 'published', 'normal', 'all', true, 89, ARRAY['мероприятие'], 7),
  ('Приём заявок на летнюю практику', 'Студенты 2-3 курсов, подайте заявки на летнюю производственную практику до 10 мая. Бланки заявок — у академического секретаря или в личном кабинете.', 'Заявки на практику до 10 мая', 'methodist@university.ru', 'published', 'normal', 'students', false, 72, ARRAY['практика'], 5),
  ('Расписание экзаменов утверждено', 'Расписание экзаменационной сессии весеннего семестра 2025/2026 утверждено приказом ректора. Ознакомиться можно в разделе документов.', 'Расписание сессии утверждено', 'secretary@university.ru', 'published', 'urgent', 'all', false, 203, ARRAY['экзамены', 'сессия'], 3),
  ('Научный семинар по ИИ', 'Кафедра информатики приглашает на научный семинар "Применение ИИ в образовании". Докладчик: доц. Иванов Д.Н. Четверг, 16:30, конференц-зал.', 'Семинар по ИИ в четверг', 'ivanov@university.ru', 'published', 'low', 'teachers', false, 34, ARRAY['семинар', 'наука'], 1),
  ('Обновление ИС (плановое)', 'В ночь с субботы на воскресенье будет проводиться плановое обновление информационной системы. Возможны кратковременные перерывы в работе.', 'Плановое обновление ИС', 'admin@university.ru', 'draft', 'normal', 'all', false, 0, ARRAY['техработы'], 0)
) AS v(title, content, summary, author_email, status, priority, audience, pinned, views, tags, days_ago)
JOIN users u ON u.email = v.author_email;

-- ============================================================================
-- 11. СОБЫТИЯ КАЛЕНДАРЯ
-- ============================================================================

INSERT INTO events (title, description, event_type, status, start_time, end_time, all_day, location, organizer_id, priority, created_at, updated_at)
SELECT v.title, v.description, v.event_type, v.status,
  (NOW() + v.start_days * INTERVAL '1 day')::timestamp,
  (NOW() + v.start_days * INTERVAL '1 day' + v.duration_hours * INTERVAL '1 hour')::timestamp,
  v.all_day, v.location, u.id, v.priority, NOW(), NOW()
FROM (VALUES
  ('Заседание кафедры информатики', 'Повестка: утверждение тем ВКР, распределение нагрузки', 'meeting', 'scheduled', 3, 2, false, 'Корпус А, к. 205', 'ivanov@university.ru', 2),
  ('День открытых дверей', 'Мероприятие для абитуриентов', 'holiday', 'scheduled', 18, 6, false, 'Конференц-зал, корпус Б', 'admin@university.ru', 3),
  ('Дедлайн: заявки на практику', 'Крайний срок подачи заявок на летнюю практику', 'deadline', 'scheduled', 13, 0, true, NULL, 'methodist@university.ru', 2),
  ('Научный семинар: ИИ в образовании', 'Доклад доц. Иванова Д.Н.', 'meeting', 'scheduled', 1, 2, false, 'Конференц-зал, корпус Б', 'ivanov@university.ru', 1),
  ('Совещание по аккредитации', 'Обсуждение хода подготовки документов', 'meeting', 'scheduled', 5, 1, false, 'Кабинет ректора', 'admin@university.ru', 3),
  ('Консультация перед экзаменом по БД', 'Консультация для ИВТ-201 перед экзаменом', 'meeting', 'scheduled', 25, 2, false, 'Корпус А, к. 301', 'ivanov@university.ru', 1)
) AS v(title, description, event_type, status, start_days, duration_hours, all_day, location, organizer_email, priority)
JOIN users u ON u.email = v.organizer_email;

-- Добавить участников к событиям
INSERT INTO event_participants (event_id, user_id, response_status, role, created_at)
SELECT e.id, u.id, v.status, v.role, NOW()
FROM (VALUES
  ('Заседание кафедры информатики', 'kuznetsov@university.ru', 'accepted', 'required'),
  ('Заседание кафедры информатики', 'methodist@university.ru', 'accepted', 'optional'),
  ('День открытых дверей', 'secretary@university.ru', 'accepted', 'required'),
  ('День открытых дверей', 'methodist@university.ru', 'pending', 'required'),
  ('Совещание по аккредитации', 'methodist@university.ru', 'accepted', 'required'),
  ('Совещание по аккредитации', 'secretary@university.ru', 'accepted', 'required'),
  ('Совещание по аккредитации', 'ivanov@university.ru', 'tentative', 'optional')
) AS v(event_title, user_email, status, role)
JOIN events e ON e.title = v.event_title
JOIN users u ON u.email = v.user_email
ON CONFLICT (event_id, user_id) DO NOTHING;

-- ============================================================================
-- 12. ОТЧЁТЫ
-- ============================================================================

INSERT INTO report_types (name, code, description, category, output_format, is_periodic, period_type) VALUES
('Успеваемость студентов', 'student_performance', 'Отчёт об успеваемости по группам', 'academic', 'xlsx', true, 'monthly'),
('Посещаемость', 'attendance', 'Отчёт о посещаемости занятий', 'academic', 'xlsx', true, 'weekly'),
('Нагрузка преподавателей', 'teacher_workload', 'Отчёт о педагогической нагрузке', 'administrative', 'pdf', true, 'quarterly'),
('Контингент студентов', 'student_count', 'Численность студентов по группам и курсам', 'administrative', 'xlsx', false, NULL)
ON CONFLICT (code) DO NOTHING;

INSERT INTO reports (report_type_id, title, description, period_start, period_end, author_id, status, data, created_at, updated_at)
SELECT rt.id, v.title, v.description, v.period_start::date, v.period_end::date, u.id, v.status,
  v.data::jsonb, NOW() - v.days_ago * INTERVAL '1 day', NOW()
FROM (VALUES
  ('attendance', 'Посещаемость ИВТ-201 за март', 'Отчёт о посещаемости группы ИВТ-201', '2026-03-01', '2026-03-31', 'secretary@university.ru', 'ready', '{"total_lessons": 48, "avg_attendance": 87.5, "students_count": 25}', 5),
  ('student_performance', 'Успеваемость 2 курса ФИТ', 'Сводный отчёт по успеваемости 2 курса', '2026-02-01', '2026-04-15', 'methodist@university.ru', 'ready', '{"avg_grade": 4.2, "excellent": 8, "good": 12, "satisfactory": 4, "unsatisfactory": 1}', 3),
  ('teacher_workload', 'Нагрузка преподавателей ФИТ Q1', 'Педагогическая нагрузка за 1 квартал', '2026-01-01', '2026-03-31', 'methodist@university.ru', 'approved', '{"teachers": 12, "total_hours": 2400, "avg_hours": 200}', 10),
  ('student_count', 'Контингент на 01.04.2026', 'Численность студентов всех направлений', '2026-04-01', '2026-04-01', 'secretary@university.ru', 'published', '{"total": 450, "by_course": {"1": 120, "2": 115, "3": 105, "4": 110}}', 1)
) AS v(type_code, title, description, period_start, period_end, author_email, status, data, days_ago)
JOIN report_types rt ON rt.code = v.type_code
JOIN users u ON u.email = v.author_email;

-- ============================================================================
-- 13. УВЕДОМЛЕНИЯ
-- ============================================================================

INSERT INTO notifications (user_id, type, priority, title, message, link, is_read, created_at, updated_at)
SELECT u.id, v.type::notification_type, v.priority::notification_priority, v.title, v.message, v.link, v.is_read,
  NOW() - v.hours_ago * INTERVAL '1 hour', NOW()
FROM (VALUES
  ('student1@university.ru', 'info', 'normal', 'Расписание обновлено', 'Расписание на текущую неделю было обновлено', '/schedule', true, 48),
  ('student1@university.ru', 'reminder', 'high', 'Экзамен через 3 дня', 'Экзамен по Базам данных состоится 30 апреля', '/schedule', false, 2),
  ('student2@university.ru', 'task', 'normal', 'Новое задание', 'Преподаватель назначил лабораторную работу №5', '/tasks', false, 6),
  ('ivanov@university.ru', 'document', 'normal', 'Документ на согласование', 'РПД по информатике ожидает вашего согласования', '/documents', false, 4),
  ('secretary@university.ru', 'system', 'normal', 'Отчёт сформирован', 'Отчёт о посещаемости за март готов к просмотру', '/reports', true, 24),
  ('methodist@university.ru', 'info', 'normal', 'Задача выполнена', 'Отчёт по посещаемости подготовлен секретарём', '/tasks', true, 12),
  ('admin@university.ru', 'warning', 'high', 'Резервное копирование', 'Последнее резервное копирование выполнено 2 дня назад', NULL, false, 1)
) AS v(user_email, type, priority, title, message, link, is_read, hours_ago)
JOIN users u ON u.email = v.user_email;

-- ============================================================================
-- 14. СООБЩЕНИЯ (messaging)
-- ============================================================================

-- Групповой чат кафедры
INSERT INTO conversations (type, title, description, created_by, created_at, updated_at)
SELECT 'group', 'Кафедра информатики', 'Рабочий чат кафедры', u.id, NOW() - INTERVAL '30 days', NOW()
FROM users u WHERE u.email = 'ivanov@university.ru';

INSERT INTO conversation_participants (conversation_id, user_id, role, joined_at)
SELECT c.id, u.id, CASE WHEN u.email = 'ivanov@university.ru' THEN 'admin' ELSE 'member' END, NOW() - INTERVAL '30 days'
FROM conversations c
JOIN users u ON u.email IN ('ivanov@university.ru', 'kuznetsov@university.ru', 'admin@university.ru')
WHERE c.title = 'Кафедра информатики'
ON CONFLICT (conversation_id, user_id) DO NOTHING;

INSERT INTO messages (conversation_id, sender_id, type, content, created_at)
SELECT c.id, u.id, 'text', v.content, NOW() - v.hours_ago * INTERVAL '1 hour'
FROM (VALUES
  ('ivanov@university.ru', 'Коллеги, не забудьте — в четверг заседание кафедры в 14:00', 72),
  ('kuznetsov@university.ru', 'Принял. Подготовлю отчёт по олимпиаде', 48),
  ('ivanov@university.ru', 'Также нужно обсудить распределение нагрузки на следующий год', 24),
  ('kuznetsov@university.ru', 'Хорошо, подготовлю предложения по часам', 12)
) AS v(sender_email, content, hours_ago)
JOIN conversations c ON c.title = 'Кафедра информатики'
JOIN users u ON u.email = v.sender_email;

-- Личный чат студент-преподаватель
INSERT INTO conversations (type, title, created_by, created_at, updated_at)
SELECT 'direct', NULL, u.id, NOW() - INTERVAL '5 days', NOW()
FROM users u WHERE u.email = 'student1@university.ru';

INSERT INTO conversation_participants (conversation_id, user_id, role, joined_at)
SELECT c.id, u.id, 'member', NOW() - INTERVAL '5 days'
FROM conversations c
JOIN users u ON u.email IN ('student1@university.ru', 'ivanov@university.ru')
WHERE c.type = 'direct' AND c.created_by = (SELECT id FROM users WHERE email = 'student1@university.ru')
ON CONFLICT (conversation_id, user_id) DO NOTHING;

INSERT INTO messages (conversation_id, sender_id, type, content, created_at)
SELECT c.id, u.id, 'text', v.content, NOW() - v.hours_ago * INTERVAL '1 hour'
FROM (VALUES
  ('student1@university.ru', 'Дмитрий Николаевич, можно сдать лабораторную на день позже? Не успеваю доделать отчёт.', 96),
  ('ivanov@university.ru', 'Добрый день, Артём. Да, можете сдать до пятницы включительно.', 90),
  ('student1@university.ru', 'Спасибо большое!', 89)
) AS v(sender_email, content, hours_ago)
JOIN conversations c ON c.type = 'direct' AND c.created_by = (SELECT id FROM users WHERE email = 'student1@university.ru')
JOIN users u ON u.email = v.sender_email;

-- ============================================================================
-- 15. ПОСЕЩАЕМОСТЬ И ОЦЕНКИ
-- ============================================================================

-- Создать запись предмета для посещаемости
INSERT INTO lessons (name, subject, teacher_id, group_name, lesson_type)
SELECT 'Базы данных — лекция', 'Базы данных', u.id, 'ИВТ-201', 'lecture'
FROM users u WHERE u.email = 'ivanov@university.ru';

INSERT INTO lessons (name, subject, teacher_id, group_name, lesson_type)
SELECT 'Базы данных — лабораторная', 'Базы данных', u.id, 'ИВТ-201', 'lab'
FROM users u WHERE u.email = 'ivanov@university.ru';

-- Посещаемость
INSERT INTO attendance_records (student_id, lesson_id, lesson_date, status, marked_by)
SELECT s.id, l.id, d::date,
  CASE WHEN random() < 0.85 THEN 'present'
       WHEN random() < 0.5 THEN 'late'
       ELSE 'absent' END,
  (SELECT id FROM users WHERE email = 'ivanov@university.ru')
FROM users s
CROSS JOIN lessons l
CROSS JOIN generate_series('2026-03-01'::date, '2026-04-25'::date, '7 days'::interval) d
WHERE s.role = 'student' AND l.subject = 'Базы данных'
ON CONFLICT (student_id, lesson_id, lesson_date) DO NOTHING;

-- Оценки
INSERT INTO grades (student_id, subject, grade_type, grade_value, graded_by, grade_date)
SELECT s.id, v.subject, v.grade_type, (55 + random() * 45)::decimal(5,2),
  (SELECT id FROM users WHERE email = v.teacher_email), v.grade_date::date
FROM users s
CROSS JOIN (VALUES
  ('Базы данных', 'homework', 'ivanov@university.ru', '2026-03-15'),
  ('Базы данных', 'midterm', 'ivanov@university.ru', '2026-04-01'),
  ('Базы данных', 'homework', 'ivanov@university.ru', '2026-04-15'),
  ('Математический анализ', 'midterm', 'kuznetsov@university.ru', '2026-04-01'),
  ('Микроэкономика', 'current', 'sidorova@university.ru', '2026-03-20')
) AS v(subject, grade_type, teacher_email, grade_date)
WHERE s.role = 'student';

-- ============================================================================
-- 16. НАСТРОЙКИ УВЕДОМЛЕНИЙ
-- ============================================================================

INSERT INTO notification_preferences (user_id, email_enabled, push_enabled, in_app_enabled, timezone)
SELECT id, true, true, true, 'Europe/Moscow' FROM users
ON CONFLICT (user_id) DO NOTHING;

COMMIT;

-- ============================================================================
-- ИТОГО: система оживлена!
-- Пользователи: 1 admin + 1 methodist + 1 secretary + 3 teachers + 4 students
-- Пароль: 12345678 (у всех)
-- Логины: admin@university.ru, methodist@university.ru, secretary@university.ru,
--          ivanov@university.ru, sidorova@university.ru, kuznetsov@university.ru,
--          student1@university.ru ... student4@university.ru
-- ============================================================================
