-- Add tsvector column for full-text search on chunks
ALTER TABLE ai_document_chunks ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Populate for existing rows
UPDATE ai_document_chunks SET search_vector = to_tsvector('russian', chunk_text)
WHERE search_vector IS NULL;

-- GIN index for full-text search
CREATE INDEX IF NOT EXISTS idx_ai_chunks_search_vector
    ON ai_document_chunks USING GIN (search_vector);

-- Trigger to auto-update search_vector on INSERT/UPDATE
CREATE OR REPLACE FUNCTION update_chunk_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector = to_tsvector('russian', NEW.chunk_text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_chunk_search_vector ON ai_document_chunks;
CREATE TRIGGER trg_chunk_search_vector
BEFORE INSERT OR UPDATE OF chunk_text ON ai_document_chunks
FOR EACH ROW EXECUTE FUNCTION update_chunk_search_vector();

-- Replace IVFFlat index with HNSW (better for small-medium datasets)
DROP INDEX IF EXISTS idx_ai_embeddings_vector;
CREATE INDEX IF NOT EXISTS idx_ai_embeddings_vector_hnsw ON ai_embeddings
    USING hnsw (embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64);

-- Reset index statuses for reindexing with new chunking parameters
UPDATE ai_document_index_status SET status = 'pending', chunks_count = 0, error_message = NULL;
