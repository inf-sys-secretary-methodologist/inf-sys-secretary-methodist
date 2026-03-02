-- Revert embedding model default to OpenAI
ALTER TABLE ai_embeddings
    ALTER COLUMN model SET DEFAULT 'text-embedding-3-small';
