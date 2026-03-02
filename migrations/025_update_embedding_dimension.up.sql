-- Update embedding model from OpenAI to Gemini (dimension stays 1536 via Matryoshka)
-- Gemini gemini-embedding-001 supports 3072/1536/768 via Matryoshka truncation
-- We use 1536 for pgvector IVFFlat compatibility (max 2000 dims)

-- Update default model name
ALTER TABLE ai_embeddings
    ALTER COLUMN model SET DEFAULT 'gemini-embedding-001';
