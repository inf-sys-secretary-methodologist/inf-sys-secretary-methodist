-- Migration: Create AI RAG Schema
-- Description: Creates tables for AI embeddings, chunks, conversations, and messages

-- Enable pgvector extension for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- AI Conversations table
CREATE TABLE IF NOT EXISTS ai_conversations (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    model VARCHAR(100) NOT NULL DEFAULT 'gpt-4o-mini',
    message_count INTEGER NOT NULL DEFAULT 0,
    last_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- AI Messages table
CREATE TABLE IF NOT EXISTS ai_messages (
    id BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
    content TEXT NOT NULL,
    tokens_used INTEGER,
    model VARCHAR(100),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Document chunks for RAG
CREATE TABLE IF NOT EXISTS ai_document_chunks (
    id BIGSERIAL PRIMARY KEY,
    document_id BIGINT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    chunk_text TEXT NOT NULL,
    chunk_tokens INTEGER,
    page_number INTEGER,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (document_id, chunk_index)
);

-- Vector embeddings for semantic search
CREATE TABLE IF NOT EXISTS ai_embeddings (
    id BIGSERIAL PRIMARY KEY,
    chunk_id BIGINT NOT NULL REFERENCES ai_document_chunks(id) ON DELETE CASCADE,
    embedding vector(1536), -- OpenAI ada-002/text-embedding-3-small dimension
    model VARCHAR(100) NOT NULL DEFAULT 'text-embedding-3-small',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (chunk_id)
);

-- AI Message sources (citations)
CREATE TABLE IF NOT EXISTS ai_message_sources (
    id BIGSERIAL PRIMARY KEY,
    message_id BIGINT NOT NULL REFERENCES ai_messages(id) ON DELETE CASCADE,
    chunk_id BIGINT NOT NULL REFERENCES ai_document_chunks(id) ON DELETE CASCADE,
    similarity_score FLOAT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Document indexing status
CREATE TABLE IF NOT EXISTS ai_document_index_status (
    document_id BIGINT PRIMARY KEY REFERENCES documents(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'indexing', 'indexed', 'failed')),
    chunks_count INTEGER DEFAULT 0,
    error_message TEXT,
    indexed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id ON ai_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_updated_at ON ai_conversations(updated_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_id ON ai_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_messages_created_at ON ai_messages(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_document_chunks_document_id ON ai_document_chunks(document_id);

CREATE INDEX IF NOT EXISTS idx_ai_message_sources_message_id ON ai_message_sources(message_id);

CREATE INDEX IF NOT EXISTS idx_ai_document_index_status_status ON ai_document_index_status(status);

-- IVFFlat index for vector similarity search (faster queries at slight accuracy cost)
-- Note: This index requires at least 1000 vectors to be effective
-- For small datasets, use exact nearest neighbor search
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_vector ON ai_embeddings
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Trigger to update conversation updated_at and message_count on new message
CREATE OR REPLACE FUNCTION update_ai_conversation_on_message()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE ai_conversations
    SET
        updated_at = NOW(),
        last_message_at = NEW.created_at,
        message_count = message_count + 1
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ai_conversation_update
AFTER INSERT ON ai_messages
FOR EACH ROW
EXECUTE FUNCTION update_ai_conversation_on_message();

-- Trigger to update index status updated_at
CREATE OR REPLACE FUNCTION update_ai_index_status_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ai_index_status_update
BEFORE UPDATE ON ai_document_index_status
FOR EACH ROW
EXECUTE FUNCTION update_ai_index_status_timestamp();

-- Add comment for documentation
COMMENT ON TABLE ai_conversations IS 'AI chat conversations for each user';
COMMENT ON TABLE ai_messages IS 'Messages in AI conversations (user questions and AI responses)';
COMMENT ON TABLE ai_document_chunks IS 'Document text chunks for RAG retrieval';
COMMENT ON TABLE ai_embeddings IS 'Vector embeddings for semantic similarity search';
COMMENT ON TABLE ai_message_sources IS 'Document sources cited in AI responses';
COMMENT ON TABLE ai_document_index_status IS 'Document indexing status tracking';
