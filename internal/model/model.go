package model

import (
	"time"
)

type Account struct {
	ID        string    `db:"id"         json:"id"`
	OwnerName string    `db:"owner_name" json:"owner_name"`
	Balance   float64   `db:"balance"    json:"balance"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Transfer struct {
	ID          string    `db:"id"           json:"id"`
	FromAccount string    `db:"from_account" json:"from_account"`
	ToAccount   string    `db:"to_account"   json:"to_account"`
	Amount      float64   `db:"amount"       json:"amount"`
	Status      string    `db:"status"       json:"status"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}

type TokenRequest struct {
	AccountID string `json:"account_id"`
	Secret    string `json:"secret"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type TransferRequest struct {
	ToAccount string  `json:"to_account"`
	Amount    float64 `json:"amount"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
