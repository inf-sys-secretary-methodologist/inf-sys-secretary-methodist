-- Migration Rollback: Drop AI RAG Schema

-- Drop triggers first
DROP TRIGGER IF EXISTS trg_ai_conversation_update ON ai_messages;
DROP TRIGGER IF EXISTS trg_ai_index_status_update ON ai_document_index_status;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_ai_conversation_on_message();
DROP FUNCTION IF EXISTS update_ai_index_status_timestamp();

-- Drop tables in reverse order of creation (respecting foreign keys)
DROP TABLE IF EXISTS ai_message_sources;
DROP TABLE IF EXISTS ai_embeddings;
DROP TABLE IF EXISTS ai_document_chunks;
DROP TABLE IF EXISTS ai_messages;
DROP TABLE IF EXISTS ai_conversations;
DROP TABLE IF EXISTS ai_document_index_status;

-- Note: We don't drop the vector extension as other tables might use it
-- DROP EXTENSION IF EXISTS vector;
