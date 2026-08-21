package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yourorg/ehailing/backend/internal/wallet/domain/entities"
)

const walletColumns = `id, user_id, balance, currency, is_platform_wallet, created_at, updated_at`

func scanWallet(rs interface{ Scan(dest ...any) error }) (*entities.Wallet, error) {
	w := &entities.Wallet{}
	if err := rs.Scan(
		&w.ID, &w.UserID, &w.Balance, &w.Currency, &w.IsPlatformWallet,
		&w.CreatedAt, &w.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return w, nil
}

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *entities.Wallet) error {
	query := `INSERT INTO wallets (` + walletColumns + `)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		wallet.ID, wallet.UserID, wallet.Balance, wallet.Currency,
		wallet.IsPlatformWallet, wallet.CreatedAt, wallet.UpdatedAt,
	)
	return err
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.Wallet, error) {
	query := `SELECT ` + walletColumns + ` FROM wallets WHERE user_id = $1`

	wallet, err := scanWallet(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return wallet, nil
}

func (r *WalletRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Wallet, error) {
	query := `SELECT ` + walletColumns + ` FROM wallets WHERE id = $1`

	wallet, err := scanWallet(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return wallet, nil
}

func (r *WalletRepository) GetPlatformWallet(ctx context.Context) (*entities.Wallet, error) {
	query := `SELECT ` + walletColumns + ` FROM wallets WHERE is_platform_wallet = TRUE LIMIT 1`

	wallet, err := scanWallet(r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, walletID uuid.UUID, newBalance float64) error {
	query := `UPDATE wallets SET balance = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, newBalance, walletID)
	return err
}
