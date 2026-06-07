package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jotagm/payment-api/internal/model"
	"github.com/jotagm/payment-api/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("payment-api/service")

type TransferService struct {
	accounts  *repository.AccountRepository
	transfers *repository.TransferRepository
}

func NewTransferService(
	accounts *repository.AccountRepository,
	transfers *repository.TransferRepository,
) *TransferService {
	return &TransferService{accounts: accounts, transfers: transfers}
}

func (s *TransferService) Execute(ctx context.Context, fromID, toID string, amount float64) (*model.Transfer, error) {
	ctx, span := tracer.Start(ctx, "TransferService.Execute")
	defer span.End()

	span.SetAttributes(
		attribute.String("transfer.from", fromID),
		attribute.String("transfer.to", toID),
		attribute.Float64("transfer.amount", amount),
	)

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	from, err := s.accounts.FindByID(ctx, fromID)
	if err != nil {
		return nil, fmt.Errorf("source account: %w", err)
	}

	if from.Balance < amount {
		return nil, errors.New("insufficient balance")
	}

	_, err = s.accounts.FindByID(ctx, toID)
	if err != nil {
		return nil, fmt.Errorf("destination account: %w", err)
	}

	tx, err := s.transfers.DB().BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.accounts.UpdateBalance(ctx, tx, fromID, -amount); err != nil {
		return nil, fmt.Errorf("debit: %w", err)
	}

	if err := s.accounts.UpdateBalance(ctx, tx, toID, amount); err != nil {
		return nil, fmt.Errorf("credit: %w", err)
	}

	transfer := &model.Transfer{
		ID:          uuid.NewString(),
		FromAccount: fromID,
		ToAccount:   toID,
		Amount:      amount,
		Status:      "completed",
	}

	if err := s.transfers.Create(ctx, tx, transfer); err != nil {
		return nil, fmt.Errorf("record transfer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	slog.InfoContext(ctx, "transfer completed",
		"transfer_id", transfer.ID,
		"from", fromID,
		"to", toID,
		"amount", amount,
	)

	return transfer, nil
}

func (s *TransferService) FindByID(ctx context.Context, id string) (*model.Transfer, error) {
	ctx, span := tracer.Start(ctx, "TransferService.FindByID")
	defer span.End()
	return s.transfers.FindByID(ctx, id)
}

type AccountService struct {
	accounts *repository.AccountRepository
}

func NewAccountService(accounts *repository.AccountRepository) *AccountService {
	return &AccountService{accounts: accounts}
}

func (s *AccountService) FindByID(ctx context.Context, id string) (*model.Account, error) {
	ctx, span := tracer.Start(ctx, "AccountService.FindByID")
	defer span.End()
	return s.accounts.FindByID(ctx, id)
}
