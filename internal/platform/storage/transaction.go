package storage

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
)

const transactionTable = "transactions"

// Transaction represents a Transaction repository.
type Transaction struct {
	storage *sqluct.Storage
}

// NewTransaction returns instance of Transaction.
func NewTransaction(storage *sqluct.Storage) *Transaction {
	return &Transaction{
		storage: storage,
	}
}

// Add adds transaction to the storage.
func (r Transaction) Add(ctx context.Context, transactionStates []model.TransactionState) error {
	errMsg := "storage.Transaction: failed to add transaction"

	transactions := make([]model.Transaction, len(transactionStates))

	for i, state := range transactionStates {
		transactions[i].TransactionState = state
	}

	q := r.storage.InsertStmt(transactionTable, transactions, sqluct.SkipZeroValues)

	_, err := r.storage.Exec(ctx, q)
	if err != nil {
		return ctxd.WrapError(
			ctx,
			err,
			errMsg,
		)
	}

	return nil
}
