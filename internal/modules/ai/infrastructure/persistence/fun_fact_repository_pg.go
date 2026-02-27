// Package persistence contains PostgreSQL implementations of AI repositories.
package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// FunFactRepositoryPg implements FunFactRepository using PostgreSQL
type FunFactRepositoryPg struct {
	db *sql.DB
}

// NewFunFactRepositoryPg creates a new FunFactRepositoryPg
func NewFunFactRepositoryPg(db *sql.DB) *FunFactRepositoryPg {
	return &FunFactRepositoryPg{db: db}
}

// Create creates a new fun fact
func (r *FunFactRepositoryPg) Create(ctx context.Context, fact *entities.FunFact) error {
	query := `INSERT INTO ai_fun_facts (content, category, source, source_url, language, is_approved)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		fact.Content, fact.Category, fact.Source, fact.SourceURL, fact.Language, fact.IsApproved,
	).Scan(&fact.ID, &fact.CreatedAt, &fact.UpdatedAt)
}

// BulkCreate creates multiple fun facts
func (r *FunFactRepositoryPg) BulkCreate(ctx context.Context, facts []entities.FunFact) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO ai_fun_facts (content, category, source, source_url, language, is_approved)
		VALUES ($1, $2, $3, $4, $5, $6)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, fact := range facts {
		_, err := stmt.ExecContext(ctx, fact.Content, fact.Category, fact.Source, fact.SourceURL, fact.Language, fact.IsApproved)
		if err != nil {
			return fmt.Errorf("failed to insert fact: %w", err)
		}
	}

	return tx.Commit()
}

// GetRandom returns a random fact, preferring least-used ones
func (r *FunFactRepositoryPg) GetRandom(ctx context.Context) (*entities.FunFact, error) {
	query := `SELECT id, content, category, source, source_url, language, is_approved,
		used_count, last_used_at, created_at, updated_at
		FROM ai_fun_facts WHERE is_approved = true
		ORDER BY used_count ASC, RANDOM() LIMIT 1`

	fact := &entities.FunFact{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&fact.ID, &fact.Content, &fact.Category, &fact.Source, &fact.SourceURL,
		&fact.Language, &fact.IsApproved, &fact.UsedCount, &fact.LastUsedAt,
		&fact.CreatedAt, &fact.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get random fact: %w", err)
	}
	return fact, nil
}

// GetLeastUsed returns the least used fact
func (r *FunFactRepositoryPg) GetLeastUsed(ctx context.Context) (*entities.FunFact, error) {
	return r.GetRandom(ctx) // Same query, both prefer least-used
}

// IncrementUsedCount increments the used count and updates last_used_at
func (r *FunFactRepositoryPg) IncrementUsedCount(ctx context.Context, id int64) error {
	query := `UPDATE ai_fun_facts SET used_count = used_count + 1, last_used_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// Count returns the total number of fun facts
func (r *FunFactRepositoryPg) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_fun_facts`).Scan(&count)
	return count, err
}
