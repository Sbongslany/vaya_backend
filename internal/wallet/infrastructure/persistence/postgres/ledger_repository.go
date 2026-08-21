package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
)

const ledgerColumns = `id, wallet_id, entry_type, amount, balance_after,
	reference_type, reference_id, description, created_by, created_at`

func scanLedgerEntry(rs interface{ Scan(dest ...any) error }) (*entities.LedgerEntry, error) {
	e := &entities.LedgerEntry{}
	if err := rs.Scan(
		&e.ID, &e.WalletID, &e.EntryType, &e.Amount, &e.BalanceAfter,
		&e.ReferenceType, &e.ReferenceID, &e.Description, &e.CreatedBy, &e.CreatedAt,
	); err != nil {
		return nil, err
	}
	return e, nil
}

type LedgerRepository struct {
	pool *pgxpool.Pool
}

func NewLedgerRepository(pool *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

func (r *LedgerRepository) Create(ctx context.Context, entry *entities.LedgerEntry) error {
	query := `INSERT INTO ledger_entries (` + ledgerColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.pool.Exec(ctx, query,
		entry.ID, entry.WalletID, entry.EntryType, entry.Amount, entry.BalanceAfter,
		entry.ReferenceType, entry.ReferenceID, entry.Description, entry.CreatedBy, entry.CreatedAt,
	)
	return err
}

func (r *LedgerRepository) FindByWalletID(ctx context.Context, walletID uuid.UUID, limit, offset int) ([]*entities.LedgerEntry, error) {
	query := `SELECT ` + ledgerColumns + ` FROM ledger_entries
		WHERE wallet_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	rows, err := r.pool.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*entities.LedgerEntry
	for rows.Next() {
		entry, err := scanLedgerEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *LedgerRepository) FindByReference(ctx context.Context, refType entities.LedgerReferenceType, refID uuid.UUID) ([]*entities.LedgerEntry, error) {
	query := `SELECT ` + ledgerColumns + ` FROM ledger_entries
		WHERE reference_type = $1 AND reference_id = $2 ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, query, refType, refID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*entities.LedgerEntry
	for rows.Next() {
		entry, err := scanLedgerEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}
