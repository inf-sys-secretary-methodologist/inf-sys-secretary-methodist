-- Remove trigger
DROP TRIGGER IF EXISTS trg_chunk_search_vector ON ai_document_chunks;
DROP FUNCTION IF EXISTS update_chunk_search_vector();

-- Remove HNSW index, restore IVFFlat
DROP INDEX IF EXISTS idx_ai_embeddings_vector_hnsw;
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_vector ON ai_embeddings
    USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- Remove search_vector column and index
DROP INDEX IF EXISTS idx_ai_chunks_search_vector;
ALTER TABLE ai_document_chunks DROP COLUMN IF EXISTS search_vector;
