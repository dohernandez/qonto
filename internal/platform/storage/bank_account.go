package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
)

const bankAccountTable = "bank_accounts"

// ErrNotEnoughBalance error represents when the account does not have enough balance.
var ErrNotEnoughBalance = errors.New("not enough balance")

// BankAccount represents a BankAccount repository.
type BankAccount struct {
	storage *sqluct.Storage

	colID               string
	colOrganizationName string
	colBalanceCents     string
	colIban             string
	colBic              string
}

// NewBankAccount returns instance of BankAccount.
func NewBankAccount(storage *sqluct.Storage) *BankAccount {
	var bankAccount model.BankAccount

	return &BankAccount{
		storage:             storage,
		colID:               storage.Mapper.Col(&bankAccount, &bankAccount.ID),
		colOrganizationName: storage.Mapper.Col(&bankAccount, &bankAccount.OrganizationName),
		colBalanceCents:     storage.Mapper.Col(&bankAccount, &bankAccount.BalanceCents),
		colIban:             storage.Mapper.Col(&bankAccount, &bankAccount.Iban),
		colBic:              storage.Mapper.Col(&bankAccount, &bankAccount.Bic),
	}
}

// BalanceCheck checks whether the account has enough balance or not from a storage.
//
// Returns the bank account detail when ever the account has enough balance, otherwise error.
func (r *BankAccount) BalanceCheck(ctx context.Context, accountState model.BankAccountState, amount model.Cents) (*model.BankAccount, error) {
	errMsg := "storage.BankAccount: failed to check account balance"

	var bankAccount model.BankAccount

	q := r.storage.SelectStmt(bankAccountTable, bankAccount).
		Where(squirrel.Eq{r.colOrganizationName: accountState.OrganizationName}).
		Where(squirrel.Eq{r.colIban: accountState.Iban}).
		Where(squirrel.Eq{r.colBic: accountState.Bic})

	if tx := sqluct.TxFromContext(ctx); tx != nil {
		q = q.Suffix("FOR UPDATE")
	}

	err := r.storage.Select(ctx, q, &bankAccount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ctxd.WrapError(ctx, ErrNotFound, errMsg)
		}

		return nil, ctxd.WrapError(
			ctx,
			err,
			errMsg,
		)
	}

	if bankAccount.BalanceCents < amount {
		return nil, ctxd.WrapError(ctx, ErrNotEnoughBalance, errMsg)
	}

	return &bankAccount, nil
}

// BalanceUpdate updates the account balance from a storage.
func (r *BankAccount) BalanceUpdate(ctx context.Context, accountID model.BankAccountID, amount model.Cents) error {
	errMsg := "storage.BankAccount: failed to update account balance"

	q := r.storage.UpdateStmt(bankAccountTable, nil).
		Set(r.colBalanceCents, amount).
		Where(squirrel.Eq{r.colID: accountID})

	if _, err := r.storage.Exec(ctx, q); err != nil {
		return ctxd.WrapError(
			ctx,
			err,
			errMsg,
		)
	}

	return nil
}
