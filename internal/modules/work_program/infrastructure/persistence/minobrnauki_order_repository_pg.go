package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/application/usecases"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/work_program/domain/repositories"
)

// moSelectColumns enumerates the minobrnauki_orders projection shared by
// GetByID and List — the order is a flat entity, so the full column set
// hydrates both paths.
const moSelectColumns = `id, order_number, title, published_at, document_id, change_scope, summary, uploaded_by, created_at`

// moListFilterClause uses cast-and-nullable predicates so a single filter
// shape works for every combination (mirror work_program / curriculum
// pattern). Empty string / sql.Null* values disable a predicate.
const moListFilterClause = `WHERE ($1 = '' OR change_scope = $1)
			AND ($2::bigint IS NULL OR uploaded_by = $2::bigint)`

// Compile-time assertion that the PG impl satisfies the wide port
// declared in application/usecases (DIP). Catches signature drift at the
// impl's compile site rather than only at DI wiring.
var _ usecases.MinobrnaukiOrderRepository = (*MinobrnaukiOrderRepositoryPG)(nil)

// MinobrnaukiOrderRepositoryPG is the SQL implementation of
// MinobrnaukiOrderRepository (приказы Минобрнауки per ADR-11). Accepts
// DBTX (not *sql.DB) so the same struct works in single-connection mode
// and against `*sql.Tx`.
type MinobrnaukiOrderRepositoryPG struct {
	db DBTX
}

// NewMinobrnaukiOrderRepositoryPG constructs the repository. db can be
// `*sql.DB` (default DI) or `*sql.Tx` (future transactional paths).
func NewMinobrnaukiOrderRepositoryPG(db DBTX) *MinobrnaukiOrderRepositoryPG {
	return &MinobrnaukiOrderRepositoryPG{db: db}
}

// Save inserts the order row plus its affected-work-program junction
// rows (minobrnauki_order_affected) atomically inside a single
// transaction. On success the generated id is written back onto the
// entity via SetID. affectedWorkProgramIDs may be empty (methodist
// records the order first, marks affected programs later per ADR-11).
// Any insert failure surfaces via fmt.Errorf wrapping and the deferred
// Rollback discards the partial state.
func (r *MinobrnaukiOrderRepositoryPG) Save(ctx context.Context, order *entities.MinobrnaukiOrder, affectedWorkProgramIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("minobrnauki_order: save: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const insertOrder = `
		INSERT INTO minobrnauki_orders (
			order_number, title, published_at, document_id,
			change_scope, summary, uploaded_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	var newID int64
	err = tx.QueryRowContext(ctx, insertOrder,
		order.OrderNumber(),
		order.Title(),
		order.PublishedAt(),
		nullableInt64Ptr(order.DocumentID()),
		string(order.ChangeScope()),
		nullableString(order.Summary()),
		order.UploadedBy(),
		order.CreatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("minobrnauki_order: save: insert order: %w", err)
	}

	const insertAffected = `INSERT INTO minobrnauki_order_affected (order_id, work_program_id) VALUES ($1, $2)`
	for _, wpID := range affectedWorkProgramIDs {
		if _, err := tx.ExecContext(ctx, insertAffected, newID, wpID); err != nil {
			return fmt.Errorf("minobrnauki_order: save: insert affected: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("minobrnauki_order: save: commit: %w", err)
	}
	order.SetID(newID)
	return nil
}

// GetByID returns the order with the given id, hydrated through
// ReconstituteMinobrnaukiOrder. Returns
// repositories.ErrMinobrnaukiOrderNotFound when no row matches. The
// affected-work-program set is fetched separately via FindAffected.
func (r *MinobrnaukiOrderRepositoryPG) GetByID(ctx context.Context, id int64) (*entities.MinobrnaukiOrder, error) {
	query := `SELECT ` + moSelectColumns + ` FROM minobrnauki_orders WHERE id = $1`

	in, err := scanOrderRow(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repositories.ErrMinobrnaukiOrderNotFound
		}
		return nil, fmt.Errorf("minobrnauki_order: get by id: %w", err)
	}
	return entities.ReconstituteMinobrnaukiOrder(in), nil
}

// List returns a page of orders matching the filter together with the
// total count of matching rows (ignoring Limit / Offset). An empty
// result is not an error.
func (r *MinobrnaukiOrderRepositoryPG) List(ctx context.Context, filter repositories.MinobrnaukiOrderListFilter) (repositories.MinobrnaukiOrderListResult, error) {
	scopeArg := ""
	if filter.ChangeScope != nil {
		scopeArg = string(*filter.ChangeScope)
	}
	var uploadedByArg sql.NullInt64
	if filter.UploadedBy != nil {
		uploadedByArg = sql.NullInt64{Int64: *filter.UploadedBy, Valid: true}
	}

	countQuery := `SELECT COUNT(*) FROM minobrnauki_orders ` + moListFilterClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, scopeArg, uploadedByArg).Scan(&total); err != nil {
		return repositories.MinobrnaukiOrderListResult{}, fmt.Errorf("minobrnauki_order: list count: %w", err)
	}

	listQuery := `SELECT ` + moSelectColumns + ` FROM minobrnauki_orders ` + moListFilterClause + `
		ORDER BY published_at DESC, id DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.QueryContext(ctx, listQuery, scopeArg, uploadedByArg, filter.Limit, filter.Offset)
	if err != nil {
		return repositories.MinobrnaukiOrderListResult{}, fmt.Errorf("minobrnauki_order: list: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []repositories.MinobrnaukiOrderListItem
	for rows.Next() {
		in, err := scanOrderRow(rows)
		if err != nil {
			return repositories.MinobrnaukiOrderListResult{}, fmt.Errorf("minobrnauki_order: list scan: %w", err)
		}
		items = append(items, repositories.MinobrnaukiOrderListItem{
			ID:          in.ID,
			OrderNumber: in.OrderNumber,
			Title:       in.Title,
			PublishedAt: in.PublishedAt,
			DocumentID:  in.DocumentID,
			ChangeScope: in.ChangeScope,
			Summary:     in.Summary,
			UploadedBy:  in.UploadedBy,
			CreatedAt:   in.CreatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return repositories.MinobrnaukiOrderListResult{}, fmt.Errorf("minobrnauki_order: list iter: %w", err)
	}
	return repositories.MinobrnaukiOrderListResult{Items: items, Total: total}, nil
}

// FindAffected returns the work_program ids linked to the given order via
// minobrnauki_order_affected, in ascending id order. An order with no
// recorded affected programs yields an empty slice (not an error).
func (r *MinobrnaukiOrderRepositoryPG) FindAffected(ctx context.Context, orderID int64) ([]int64, error) {
	const query = `SELECT work_program_id FROM minobrnauki_order_affected WHERE order_id = $1 ORDER BY work_program_id`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("minobrnauki_order: find affected: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("minobrnauki_order: find affected scan: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("minobrnauki_order: find affected iter: %w", err)
	}
	return ids, nil
}

// rowScanner is the narrow surface shared by *sql.Row and *sql.Rows so
// scanOrderRow serves both GetByID (single row) and List (iteration).
type rowScanner interface {
	Scan(dest ...any) error
}

// scanOrderRow scans one minobrnauki_orders row (in moSelectColumns
// order) into a Reconstitute input, unwrapping nullable columns
// (document_id, summary) into their pointer / string zero values.
func scanOrderRow(s rowScanner) (entities.ReconstituteMinobrnaukiOrderInput, error) {
	var (
		id, uploadedBy                  int64
		orderNumber, title, changeScope string
		publishedAt, createdAt          time.Time
		documentID                      sql.NullInt64
		summary                         sql.NullString
	)
	if err := s.Scan(
		&id, &orderNumber, &title, &publishedAt, &documentID,
		&changeScope, &summary, &uploadedBy, &createdAt,
	); err != nil {
		return entities.ReconstituteMinobrnaukiOrderInput{}, err
	}

	in := entities.ReconstituteMinobrnaukiOrderInput{
		ID:          id,
		OrderNumber: orderNumber,
		Title:       title,
		PublishedAt: publishedAt,
		ChangeScope: domain.MinobrnaukiOrderChangeScope(changeScope),
		UploadedBy:  uploadedBy,
		CreatedAt:   createdAt,
	}
	if documentID.Valid {
		v := documentID.Int64
		in.DocumentID = &v
	}
	if summary.Valid {
		in.Summary = summary.String
	}
	return in, nil
}
