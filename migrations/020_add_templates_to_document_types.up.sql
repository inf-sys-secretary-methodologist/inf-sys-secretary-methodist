-- ============================================================================
-- DOCUMENT TEMPLATES - Add template support to document_types
-- ============================================================================

-- Add template columns to document_types
ALTER TABLE document_types
    ADD COLUMN IF NOT EXISTS template_content TEXT,
    ADD COLUMN IF NOT EXISTS template_variables JSONB;

-- Comment on columns
COMMENT ON COLUMN document_types.template_content IS 'Template content with placeholders like {{variable_name}}';
COMMENT ON COLUMN document_types.template_variables IS 'JSON array defining template variables: [{name, label, type, required, default_value, options}]';

-- Add sample templates to existing document types
UPDATE document_types SET
    template_content = 'СЛУЖЕБНАЯ ЗАПИСКА

Кому: {{recipient_position}} {{recipient_name}}
От: {{author_position}} {{author_name}}
Дата: {{date}}

О: {{subject}}

{{content}}

С уважением,
{{author_name}}
{{author_position}}',
    template_variables = '[
        {"name": "recipient_position", "label": "Должность получателя", "type": "string", "required": true},
        {"name": "recipient_name", "label": "ФИО получателя", "type": "string", "required": true},
        {"name": "author_position", "label": "Должность автора", "type": "string", "required": true},
        {"name": "author_name", "label": "ФИО автора", "type": "string", "required": true},
        {"name": "date", "label": "Дата", "type": "date", "required": true},
        {"name": "subject", "label": "Тема", "type": "string", "required": true},
        {"name": "content", "label": "Содержание", "type": "text", "required": true}
    ]'::jsonb
WHERE code = 'memo';

UPDATE document_types SET
    template_content = 'ПРИКАЗ № {{number}}

от {{date}}

{{title}}

В соответствии с {{basis}}

ПРИКАЗЫВАЮ:

{{content}}

{{director_position}}                                  {{director_name}}',
    template_variables = '[
        {"name": "number", "label": "Номер приказа", "type": "string", "required": true},
        {"name": "date", "label": "Дата", "type": "date", "required": true},
        {"name": "title", "label": "Заголовок", "type": "string", "required": true},
        {"name": "basis", "label": "Основание", "type": "text", "required": false},
        {"name": "content", "label": "Текст приказа", "type": "text", "required": true},
        {"name": "director_position", "label": "Должность руководителя", "type": "string", "required": true},
        {"name": "director_name", "label": "ФИО руководителя", "type": "string", "required": true}
    ]'::jsonb
WHERE code IN ('order_main', 'order_hr', 'order_admin');

UPDATE document_types SET
    template_content = 'ДЕЛОВОЕ ПИСЬМО

{{recipient_organization}}
{{recipient_address}}

{{recipient_name}}

Уважаемый(ая) {{recipient_greeting}}!

{{content}}

С уважением,
{{author_name}}
{{author_position}}
{{author_organization}}

Тел: {{author_phone}}
Email: {{author_email}}',
    template_variables = '[
        {"name": "recipient_organization", "label": "Организация получателя", "type": "string", "required": false},
        {"name": "recipient_address", "label": "Адрес получателя", "type": "text", "required": false},
        {"name": "recipient_name", "label": "ФИО получателя", "type": "string", "required": true},
        {"name": "recipient_greeting", "label": "Обращение", "type": "string", "required": true},
        {"name": "content", "label": "Текст письма", "type": "text", "required": true},
        {"name": "author_name", "label": "ФИО автора", "type": "string", "required": true},
        {"name": "author_position", "label": "Должность автора", "type": "string", "required": true},
        {"name": "author_organization", "label": "Организация", "type": "string", "required": true},
        {"name": "author_phone", "label": "Телефон", "type": "string", "required": false},
        {"name": "author_email", "label": "Email", "type": "string", "required": false}
    ]'::jsonb
WHERE code = 'business_letter';

UPDATE document_types SET
    template_content = 'ПРОТОКОЛ № {{number}}

заседания {{meeting_type}}
от {{date}}

Присутствовали: {{attendees}}

Повестка дня:
{{agenda}}

{{content}}

Председатель: {{chairman_name}}
Секретарь: {{secretary_name}}',
    template_variables = '[
        {"name": "number", "label": "Номер протокола", "type": "string", "required": true},
        {"name": "meeting_type", "label": "Тип заседания", "type": "string", "required": true},
        {"name": "date", "label": "Дата", "type": "date", "required": true},
        {"name": "attendees", "label": "Присутствующие", "type": "text", "required": true},
        {"name": "agenda", "label": "Повестка дня", "type": "text", "required": true},
        {"name": "content", "label": "Ход заседания и решения", "type": "text", "required": true},
        {"name": "chairman_name", "label": "ФИО председателя", "type": "string", "required": true},
        {"name": "secretary_name", "label": "ФИО секретаря", "type": "string", "required": true}
    ]'::jsonb
WHERE code = 'protocol';

-- Create index for searching templates
CREATE INDEX IF NOT EXISTS idx_document_types_has_template
    ON document_types ((template_content IS NOT NULL));
