-- AI Fun Facts table for educational content delivery
CREATE TABLE IF NOT EXISTS ai_fun_facts (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'education',
    source VARCHAR(500),
    source_url VARCHAR(1000),
    language VARCHAR(10) NOT NULL DEFAULT 'ru',
    is_approved BOOLEAN NOT NULL DEFAULT true,
    used_count INTEGER NOT NULL DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_fun_facts_category ON ai_fun_facts(category);
CREATE INDEX idx_ai_fun_facts_used_count ON ai_fun_facts(used_count ASC);
CREATE INDEX idx_ai_fun_facts_language ON ai_fun_facts(language);
