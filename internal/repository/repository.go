package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jotagm/payment-api/internal/model"
)

type AccountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*model.Account, error) {
	var account model.Account
	err := r.db.GetContext(ctx, &account, "SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}
	return &account, nil
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, tx *sqlx.Tx, id string, delta float64) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE accounts SET balance = balance + $1 WHERE id = $2",
		delta, id,
	)
	return err
}

type TransferRepository struct {
	db *sqlx.DB
}

func NewTransferRepository(db *sqlx.DB) *TransferRepository {
	return &TransferRepository{db: db}
}

func (r *TransferRepository) Create(ctx context.Context, tx *sqlx.Tx, t *model.Transfer) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO transfers (id, from_account, to_account, amount, status)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID, t.FromAccount, t.ToAccount, t.Amount, t.Status,
	)
	return err
}

func (r *TransferRepository) FindByID(ctx context.Context, id string) (*model.Transfer, error) {
	var t model.Transfer
	err := r.db.GetContext(ctx, &t, "SELECT * FROM transfers WHERE id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("transfer not found: %w", err)
	}
	return &t, nil
}

func (r *TransferRepository) DB() *sqlx.DB {
	return r.db
}
