-- ============================================================================
-- STUDENT_DEBTS — академические долги студентов (модуль student_debts)
-- ============================================================================
-- Initiative: docs/plans/2026-06-24-student-debts-design.md (PR 2 of 6).
--
-- Bounded context: один студент × одна дисциплина × один семестр = один долг.
-- Источник истины — за портом DebtImporter/Exporter (Excel сейчас, 1С позже),
-- поэтому identity-поля денормализованы из источника, а связи на внутренние
-- сущности (users, curriculum_section_items) — best-effort и nullable.
--
-- Defense-in-depth: каждый доменный инвариант из
-- internal/modules/student_debts/domain/entities/*.go продублирован CHECK-ом.
-- Домен валидирует на записи; CHECK ловит прямой SQL и Reconstitute-путь.
-- Enum-значения CHECK совпадают byte-for-byte с доменными константами
-- (DebtStatus / ControlForm / ResitResult).
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1) student_debts — корень агрегата
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS student_debts (
    id                BIGSERIAL    PRIMARY KEY,

    -- Денормализовано из источника импорта (Excel/1С).
    student_full_name VARCHAR(255) NOT NULL,
    group_name        VARCHAR(100) NOT NULL,
    discipline_name   VARCHAR(255) NOT NULL,
    semester          INT          NOT NULL,
    control_form      VARCHAR(32)  NOT NULL,

    -- Best-effort связи: проставляются, когда студент/дисциплина существуют
    -- локально. ON DELETE SET NULL — удаление пользователя/дисциплины НЕ
    -- стирает академическую запись о долге.
    student_user_id   BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    discipline_id     BIGINT       REFERENCES curriculum_section_items(id) ON DELETE SET NULL,

    -- Провенанс импорта / идемпотентность ре-импорта.
    source_ref        VARCHAR(255) NOT NULL DEFAULT '',
    source_hash       VARCHAR(64)  NOT NULL DEFAULT '',

    status            VARCHAR(32)  NOT NULL DEFAULT 'open',
    version           INT          NOT NULL DEFAULT 1,

    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_sd_student_name_nonempty
        CHECK (length(trim(student_full_name)) > 0),
    CONSTRAINT chk_sd_group_nonempty
        CHECK (length(trim(group_name)) > 0),
    CONSTRAINT chk_sd_discipline_nonempty
        CHECK (length(trim(discipline_name)) > 0),
    -- Mirrors NewStudentDebt: semester ∈ [1,12].
    CONSTRAINT chk_sd_semester_range
        CHECK (semester BETWEEN 1 AND 12),
    -- Mirrors ControlForm constants byte-for-byte.
    CONSTRAINT chk_sd_control_form_enum
        CHECK (control_form IN ('zachet','exam','course_project','differential_zachet')),
    -- Mirrors DebtStatus constants byte-for-byte.
    CONSTRAINT chk_sd_status_enum
        CHECK (status IN ('open','resit_scheduled','commission','closed_passed','closed_failed')),
    CONSTRAINT chk_sd_version_positive
        CHECK (version >= 1),

    -- Натуральный ключ долга — основа идемпотентного upsert при ре-импорте
    -- (строка без служебного ID матчится по этому ключу).
    CONSTRAINT uq_student_debts_identity
        UNIQUE (group_name, student_full_name, discipline_name, semester)
);

CREATE INDEX IF NOT EXISTS idx_student_debts_student_user_id ON student_debts(student_user_id);
CREATE INDEX IF NOT EXISTS idx_student_debts_group_name      ON student_debts(group_name);
CREATE INDEX IF NOT EXISTS idx_student_debts_status          ON student_debts(status);
CREATE INDEX IF NOT EXISTS idx_student_debts_discipline_id   ON student_debts(discipline_id);
CREATE INDEX IF NOT EXISTS idx_student_debts_semester        ON student_debts(semester);

COMMENT ON TABLE  student_debts IS 'Академические долги студентов; источник истины за портом DebtImporter (Excel/1С)';
COMMENT ON COLUMN student_debts.student_user_id IS 'Best-effort FK на users; NULL если студент не сопоставлен локально';
COMMENT ON COLUMN student_debts.discipline_id   IS 'Best-effort FK на curriculum_section_items; NULL если дисциплина не сопоставлена';
COMMENT ON COLUMN student_debts.source_hash     IS 'SHA-256 строки источника — детект изменений при ре-импорте (idempotency)';
COMMENT ON COLUMN student_debts.version         IS 'Оптимистическая блокировка для конкурентных правок методистов';

-- ----------------------------------------------------------------------------
-- 2) debt_resit_attempts — попытки пересдачи (child entity)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS debt_resit_attempts (
    id             BIGSERIAL   PRIMARY KEY,
    debt_id        BIGINT      NOT NULL REFERENCES student_debts(id) ON DELETE CASCADE,
    attempt_no     INT         NOT NULL,
    scheduled_date TIMESTAMPTZ NOT NULL,
    examiner       VARCHAR(255) NOT NULL,
    is_commission  BOOLEAN     NOT NULL DEFAULT FALSE,
    result         VARCHAR(16) NOT NULL DEFAULT 'pending',
    grade          INT,
    recorded_by    BIGINT      REFERENCES users(id) ON DELETE SET NULL,
    recorded_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Mirrors NewResitAttempt: attempt number monotonic & positive.
    CONSTRAINT chk_dra_attempt_no_positive
        CHECK (attempt_no > 0),
    CONSTRAINT chk_dra_examiner_nonempty
        CHECK (length(trim(examiner)) > 0),
    -- Mirrors ResitResult constants byte-for-byte.
    CONSTRAINT chk_dra_result_enum
        CHECK (result IN ('pending','passed','failed','no_show')),
    -- One attempt number per debt — mirrors the domain's monotonic AttemptNo.
    CONSTRAINT uq_debt_resit_attempts_no
        UNIQUE (debt_id, attempt_no)
);

CREATE INDEX IF NOT EXISTS idx_debt_resit_attempts_debt_id ON debt_resit_attempts(debt_id);

COMMENT ON TABLE  debt_resit_attempts IS 'Попытки пересдачи долга (обычные + комиссионная); child агрегата student_debts';
COMMENT ON COLUMN debt_resit_attempts.is_commission IS 'TRUE — комиссионная пересдача (последняя попытка после исчерпания обычных)';

-- ----------------------------------------------------------------------------
-- 3) updated_at maintenance trigger
-- ----------------------------------------------------------------------------
-- Домен (NewStudentDebt) не проставляет created_at/updated_at — таймстемпы
-- ведёт БД: DEFAULT NOW() на вставке + триггер на обновлении. Repo не пишет
-- updated_at явно; чтение ре-гидратирует значение из БД.
CREATE OR REPLACE FUNCTION update_student_debts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_student_debts_updated_at
    BEFORE UPDATE ON student_debts
    FOR EACH ROW
    EXECUTE FUNCTION update_student_debts_updated_at();
