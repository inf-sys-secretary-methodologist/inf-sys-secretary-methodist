// Package persistence contains repository implementations for the AI module.
package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/repositories"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// EmbeddingRepositoryPg implements EmbeddingRepository using PostgreSQL with pgvector
type EmbeddingRepositoryPg struct {
	db *sql.DB
}

// NewEmbeddingRepositoryPg creates a new PostgreSQL embedding repository
func NewEmbeddingRepositoryPg(db *sql.DB) repositories.EmbeddingRepository {
	return &EmbeddingRepositoryPg{db: db}
}

// CreateChunk creates a document chunk
func (r *EmbeddingRepositoryPg) CreateChunk(ctx context.Context, chunk *entities.DocumentChunk) error {
	query := `
		INSERT INTO ai_document_chunks (document_id, chunk_index, chunk_text, chunk_tokens, page_number, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	metadata, _ := json.Marshal(chunk.Metadata)
	chunk.CreatedAt = time.Now()

	return r.db.QueryRowContext(
		ctx, query,
		chunk.DocumentID,
		chunk.ChunkIndex,
		chunk.ChunkText,
		chunk.ChunkTokens,
		chunk.PageNumber,
		metadata,
		chunk.CreatedAt,
	).Scan(&chunk.ID)
}

// CreateChunks creates multiple document chunks
func (r *EmbeddingRepositoryPg) CreateChunks(ctx context.Context, chunks []entities.DocumentChunk) error {
	if len(chunks) == 0 {
		return nil
	}

	query := `
		INSERT INTO ai_document_chunks (document_id, chunk_index, chunk_text, chunk_tokens, page_number, metadata, created_at)
		VALUES `

	values := make([]string, 0, len(chunks))
	args := make([]interface{}, 0, len(chunks)*7)
	now := time.Now()

	for i, chunk := range chunks {
		offset := i * 7
		values = append(values, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7))

		metadata, _ := json.Marshal(chunk.Metadata)
		args = append(args, chunk.DocumentID, chunk.ChunkIndex, chunk.ChunkText,
			chunk.ChunkTokens, chunk.PageNumber, metadata, now)
	}

	query += strings.Join(values, ", ") // #nosec G202 -- parameterized placeholders, not user input
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// GetChunksByDocumentID retrieves chunks for a document
func (r *EmbeddingRepositoryPg) GetChunksByDocumentID(ctx context.Context, documentID int64) ([]entities.DocumentChunk, error) {
	query := `
		SELECT id, document_id, chunk_index, chunk_text, chunk_tokens, page_number, metadata, created_at
		FROM ai_document_chunks
		WHERE document_id = $1
		ORDER BY chunk_index`

	rows, err := r.db.QueryContext(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query chunks: %w", err)
	}
	defer rows.Close()

	chunks := make([]entities.DocumentChunk, 0)
	for rows.Next() {
		var c entities.DocumentChunk
		var metadata []byte
		if err := rows.Scan(
			&c.ID,
			&c.DocumentID,
			&c.ChunkIndex,
			&c.ChunkText,
			&c.ChunkTokens,
			&c.PageNumber,
			&metadata,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan chunk: %w", err)
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &c.Metadata)
		}
		chunks = append(chunks, c)
	}

	return chunks, nil
}

// DeleteChunksByDocumentID deletes all chunks for a document
func (r *EmbeddingRepositoryPg) DeleteChunksByDocumentID(ctx context.Context, documentID int64) error {
	// First delete embeddings
	embeddingQuery := `
		DELETE FROM ai_embeddings
		WHERE chunk_id IN (SELECT id FROM ai_document_chunks WHERE document_id = $1)`
	if _, err := r.db.ExecContext(ctx, embeddingQuery, documentID); err != nil {
		return fmt.Errorf("failed to delete embeddings: %w", err)
	}

	// Then delete chunks
	query := `DELETE FROM ai_document_chunks WHERE document_id = $1`
	_, err := r.db.ExecContext(ctx, query, documentID)
	return err
}

// CreateEmbedding creates an embedding for a chunk
func (r *EmbeddingRepositoryPg) CreateEmbedding(ctx context.Context, embedding *entities.Embedding) error {
	query := `
		INSERT INTO ai_embeddings (chunk_id, embedding, model, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chunk_id) DO UPDATE SET embedding = $2, model = $3, created_at = $4
		RETURNING id`

	embedding.CreatedAt = time.Now()
	vec := pgvector.NewVector(embedding.Embedding)

	return r.db.QueryRowContext(
		ctx, query,
		embedding.ChunkID,
		vec,
		embedding.Model,
		embedding.CreatedAt,
	).Scan(&embedding.ID)
}

// CreateEmbeddings creates multiple embeddings
func (r *EmbeddingRepositoryPg) CreateEmbeddings(ctx context.Context, embeddings []entities.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	// Use batch insert with ON CONFLICT
	query := `
		INSERT INTO ai_embeddings (chunk_id, embedding, model, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chunk_id) DO UPDATE SET embedding = EXCLUDED.embedding, model = EXCLUDED.model, created_at = EXCLUDED.created_at`

	now := time.Now()
	for _, embedding := range embeddings {
		vec := pgvector.NewVector(embedding.Embedding)
		if _, err := r.db.ExecContext(ctx, query, embedding.ChunkID, vec, embedding.Model, now); err != nil {
			return fmt.Errorf("failed to create embedding: %w", err)
		}
	}

	return nil
}

// SearchSimilar finds similar chunks using vector similarity
func (r *EmbeddingRepositoryPg) SearchSimilar(ctx context.Context, embedding []float32, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	query := `
		SELECT
			c.id, c.document_id, c.chunk_index, c.chunk_text, c.chunk_tokens, c.page_number, c.metadata, c.created_at,
			d.title as document_title,
			1 - (e.embedding <=> $1) as similarity
		FROM ai_embeddings e
		JOIN ai_document_chunks c ON e.chunk_id = c.id
		JOIN documents d ON c.document_id = d.id
		WHERE 1 - (e.embedding <=> $1) >= $2
		ORDER BY similarity DESC
		LIMIT $3`

	vec := pgvector.NewVector(embedding)
	rows, err := r.db.QueryContext(ctx, query, vec, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar: %w", err)
	}
	defer rows.Close()

	results := make([]entities.ChunkWithScore, 0)
	for rows.Next() {
		var result entities.ChunkWithScore
		result.Chunk = &entities.DocumentChunk{}
		var metadata []byte

		if err := rows.Scan(
			&result.Chunk.ID,
			&result.Chunk.DocumentID,
			&result.Chunk.ChunkIndex,
			&result.Chunk.ChunkText,
			&result.Chunk.ChunkTokens,
			&result.Chunk.PageNumber,
			&metadata,
			&result.Chunk.CreatedAt,
			&result.DocumentTitle,
			&result.SimilarityScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &result.Chunk.Metadata)
		}

		results = append(results, result)
	}

	return results, nil
}

// SearchSimilarByDocumentTypes finds similar chunks filtered by document types
func (r *EmbeddingRepositoryPg) SearchSimilarByDocumentTypes(ctx context.Context, embedding []float32, docTypes []string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	query := `
		SELECT
			c.id, c.document_id, c.chunk_index, c.chunk_text, c.chunk_tokens, c.page_number, c.metadata, c.created_at,
			d.title as document_title,
			1 - (e.embedding <=> $1) as similarity
		FROM ai_embeddings e
		JOIN ai_document_chunks c ON e.chunk_id = c.id
		JOIN documents d ON c.document_id = d.id
		WHERE 1 - (e.embedding <=> $1) >= $2 AND d.category = ANY($3)
		ORDER BY similarity DESC
		LIMIT $4`

	vec := pgvector.NewVector(embedding)
	rows, err := r.db.QueryContext(ctx, query, vec, threshold, pq.Array(docTypes), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search similar: %w", err)
	}
	defer rows.Close()

	results := make([]entities.ChunkWithScore, 0)
	for rows.Next() {
		var result entities.ChunkWithScore
		result.Chunk = &entities.DocumentChunk{}
		var metadata []byte

		if err := rows.Scan(
			&result.Chunk.ID,
			&result.Chunk.DocumentID,
			&result.Chunk.ChunkIndex,
			&result.Chunk.ChunkText,
			&result.Chunk.ChunkTokens,
			&result.Chunk.PageNumber,
			&metadata,
			&result.Chunk.CreatedAt,
			&result.DocumentTitle,
			&result.SimilarityScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &result.Chunk.Metadata)
		}

		results = append(results, result)
	}

	return results, nil
}

// GetIndexStatus retrieves the indexing status for a document
func (r *EmbeddingRepositoryPg) GetIndexStatus(ctx context.Context, documentID int64) (*entities.DocumentIndexStatus, error) {
	query := `
		SELECT document_id, status, chunks_count, error_message, indexed_at, created_at, updated_at
		FROM ai_document_index_status
		WHERE document_id = $1`

	status := &entities.DocumentIndexStatus{}
	err := r.db.QueryRowContext(ctx, query, documentID).Scan(
		&status.DocumentID,
		&status.Status,
		&status.ChunksCount,
		&status.ErrorMessage,
		&status.IndexedAt,
		&status.CreatedAt,
		&status.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get index status: %w", err)
	}

	return status, nil
}

// SetIndexStatus sets the indexing status for a document
func (r *EmbeddingRepositoryPg) SetIndexStatus(ctx context.Context, status *entities.DocumentIndexStatus) error {
	query := `
		INSERT INTO ai_document_index_status (document_id, status, chunks_count, error_message, indexed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (document_id) DO UPDATE SET
			status = $2,
			chunks_count = $3,
			error_message = $4,
			indexed_at = CASE WHEN $2 = 'indexed' THEN NOW() ELSE ai_document_index_status.indexed_at END,
			updated_at = NOW()`

	now := time.Now()
	status.CreatedAt = now
	status.UpdatedAt = now

	var indexedAt *time.Time
	if status.Status == entities.IndexStatusIndexed {
		indexedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query,
		status.DocumentID,
		status.Status,
		status.ChunksCount,
		status.ErrorMessage,
		indexedAt,
		status.CreatedAt,
		status.UpdatedAt,
	)
	return err
}

// GetPendingDocuments retrieves documents that need indexing
func (r *EmbeddingRepositoryPg) GetPendingDocuments(ctx context.Context, limit int) ([]int64, error) {
	// Get documents that don't have index status or have pending/failed status
	query := `
		SELECT d.id
		FROM documents d
		LEFT JOIN ai_document_index_status s ON d.id = s.document_id
		WHERE s.document_id IS NULL OR s.status IN ('pending', 'failed')
		ORDER BY d.created_at DESC
		LIMIT $1`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending documents: %w", err)
	}
	defer rows.Close()

	ids := make([]int64, 0)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan document id: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// GetIndexingStats retrieves indexing statistics
func (r *EmbeddingRepositoryPg) GetIndexingStats(ctx context.Context) (totalDocuments, indexedDocuments, pendingDocuments int, lastIndexedAt *string, err error) {
	// Total documents
	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM documents`).Scan(&totalDocuments); err != nil {
		return 0, 0, 0, nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Indexed documents
	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_document_index_status WHERE status = 'indexed'`).Scan(&indexedDocuments); err != nil {
		return 0, 0, 0, nil, fmt.Errorf("failed to count indexed documents: %w", err)
	}

	// Pending documents
	query := `
		SELECT COUNT(*)
		FROM documents d
		LEFT JOIN ai_document_index_status s ON d.id = s.document_id
		WHERE s.document_id IS NULL OR s.status IN ('pending', 'failed')`
	if err = r.db.QueryRowContext(ctx, query).Scan(&pendingDocuments); err != nil {
		return 0, 0, 0, nil, fmt.Errorf("failed to count pending documents: %w", err)
	}

	// Last indexed at
	var lastIndexed sql.NullTime
	if err = r.db.QueryRowContext(ctx, `SELECT MAX(indexed_at) FROM ai_document_index_status WHERE status = 'indexed'`).Scan(&lastIndexed); err != nil {
		return 0, 0, 0, nil, fmt.Errorf("failed to get last indexed: %w", err)
	}

	if lastIndexed.Valid {
		formatted := lastIndexed.Time.Format(time.RFC3339)
		lastIndexedAt = &formatted
	}

	return totalDocuments, indexedDocuments, pendingDocuments, lastIndexedAt, nil
}

// SearchHybrid performs hybrid search combining vector similarity and full-text search using RRF.
func (r *EmbeddingRepositoryPg) SearchHybrid(ctx context.Context, embedding []float32, queryText string, limit int, threshold float64) ([]entities.ChunkWithScore, error) {
	query := `
		WITH vector_results AS (
			SELECT c.id, ROW_NUMBER() OVER (ORDER BY e.embedding <=> $1) as rank
			FROM ai_embeddings e
			JOIN ai_document_chunks c ON e.chunk_id = c.id
			WHERE 1 - (e.embedding <=> $1) >= $2
			LIMIT $3 * 3
		),
		fts_results AS (
			SELECT c.id, ROW_NUMBER() OVER (ORDER BY ts_rank(c.search_vector, plainto_tsquery('russian', $4)) DESC) as rank
			FROM ai_document_chunks c
			WHERE c.search_vector @@ plainto_tsquery('russian', $4)
			LIMIT $3 * 3
		),
		combined AS (
			SELECT COALESCE(v.id, f.id) as chunk_id,
				   1.0/(60 + COALESCE(v.rank, 1000)) + 1.0/(60 + COALESCE(f.rank, 1000)) as rrf_score
			FROM vector_results v
			FULL OUTER JOIN fts_results f ON v.id = f.id
		)
		SELECT c.id, c.document_id, c.chunk_index, c.chunk_text, c.chunk_tokens, c.page_number, c.metadata, c.created_at,
			   d.title as document_title,
			   cm.rrf_score as similarity
		FROM combined cm
		JOIN ai_document_chunks c ON cm.chunk_id = c.id
		JOIN documents d ON c.document_id = d.id
		ORDER BY cm.rrf_score DESC
		LIMIT $3`

	vec := pgvector.NewVector(embedding)
	rows, err := r.db.QueryContext(ctx, query, vec, threshold, limit, queryText)
	if err != nil {
		return nil, fmt.Errorf("failed to search hybrid: %w", err)
	}
	defer rows.Close()

	results := make([]entities.ChunkWithScore, 0)
	for rows.Next() {
		var result entities.ChunkWithScore
		result.Chunk = &entities.DocumentChunk{}
		var metadata []byte

		if err := rows.Scan(
			&result.Chunk.ID,
			&result.Chunk.DocumentID,
			&result.Chunk.ChunkIndex,
			&result.Chunk.ChunkText,
			&result.Chunk.ChunkTokens,
			&result.Chunk.PageNumber,
			&metadata,
			&result.Chunk.CreatedAt,
			&result.DocumentTitle,
			&result.SimilarityScore,
		); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &result.Chunk.Metadata)
		}

		results = append(results, result)
	}

	return results, nil
}

// GetAdjacentChunks retrieves neighboring chunks (±windowSize) for the given chunk IDs.
func (r *EmbeddingRepositoryPg) GetAdjacentChunks(ctx context.Context, chunkIDs []int64, windowSize int) ([]entities.DocumentChunk, error) {
	if len(chunkIDs) == 0 || windowSize <= 0 {
		return nil, nil
	}

	query := `
		SELECT DISTINCT ac.id, ac.document_id, ac.chunk_index, ac.chunk_text, ac.chunk_tokens, ac.page_number, ac.metadata, ac.created_at
		FROM ai_document_chunks c
		JOIN ai_document_chunks ac ON ac.document_id = c.document_id
			AND ac.chunk_index BETWEEN c.chunk_index - $2 AND c.chunk_index + $2
			AND ac.id != c.id
		WHERE c.id = ANY($1)
		ORDER BY ac.document_id, ac.chunk_index`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(chunkIDs), windowSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get adjacent chunks: %w", err)
	}
	defer rows.Close()

	chunks := make([]entities.DocumentChunk, 0)
	for rows.Next() {
		var c entities.DocumentChunk
		var metadata []byte
		if err := rows.Scan(
			&c.ID,
			&c.DocumentID,
			&c.ChunkIndex,
			&c.ChunkText,
			&c.ChunkTokens,
			&c.PageNumber,
			&metadata,
			&c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan adjacent chunk: %w", err)
		}
		if len(metadata) > 0 {
			_ = json.Unmarshal(metadata, &c.Metadata)
		}
		chunks = append(chunks, c)
	}

	return chunks, nil
}
