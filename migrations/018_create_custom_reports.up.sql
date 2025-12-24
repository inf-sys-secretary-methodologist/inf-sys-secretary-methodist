-- Create custom_reports table for custom report builder
-- This table stores user-defined report templates with dynamic fields and filters

CREATE TABLE IF NOT EXISTS custom_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    data_source VARCHAR(50) NOT NULL CHECK (data_source IN ('documents', 'users', 'events', 'tasks', 'students')),
    fields JSONB NOT NULL DEFAULT '[]'::jsonb,
    filters JSONB NOT NULL DEFAULT '[]'::jsonb,
    groupings JSONB NOT NULL DEFAULT '[]'::jsonb,
    sortings JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE
);

-- Indexes for common queries
CREATE INDEX idx_custom_reports_created_by ON custom_reports(created_by);
CREATE INDEX idx_custom_reports_data_source ON custom_reports(data_source);
CREATE INDEX idx_custom_reports_is_public ON custom_reports(is_public);
CREATE INDEX idx_custom_reports_updated_at ON custom_reports(updated_at DESC);
CREATE INDEX idx_custom_reports_name_search ON custom_reports USING gin(to_tsvector('simple', name));

-- Trigger to update updated_at on changes
CREATE OR REPLACE FUNCTION update_custom_reports_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_custom_reports_updated_at
    BEFORE UPDATE ON custom_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_custom_reports_updated_at();

-- Comments for documentation
COMMENT ON TABLE custom_reports IS 'User-defined report templates for custom report builder';
COMMENT ON COLUMN custom_reports.id IS 'Unique identifier (UUID)';
COMMENT ON COLUMN custom_reports.name IS 'Report template name';
COMMENT ON COLUMN custom_reports.description IS 'Optional description of the report';
COMMENT ON COLUMN custom_reports.data_source IS 'Data source type (documents, users, events, tasks, students)';
COMMENT ON COLUMN custom_reports.fields IS 'Selected fields configuration as JSON array';
COMMENT ON COLUMN custom_reports.filters IS 'Filter conditions as JSON array';
COMMENT ON COLUMN custom_reports.groupings IS 'Grouping configuration as JSON array';
COMMENT ON COLUMN custom_reports.sortings IS 'Sorting configuration as JSON array';
COMMENT ON COLUMN custom_reports.created_by IS 'User ID who created the report';
COMMENT ON COLUMN custom_reports.is_public IS 'Whether the report is publicly accessible';
