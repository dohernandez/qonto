package usecase

import (
	"context"

	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
)

// TransactionBulk defines the functionality of the use case TransactionBulk used to process a transaction in bulk.
type TransactionBulk interface {
	// TransactionBulk use case functionality to process a transaction in bulk.
	TransactionBulk(ctx context.Context, input TransactionBulkInput) error
}

// TransactionBulkInput contains all the inputs require executing TransactionBulk use case.
type TransactionBulkInput struct {
	OrganizationName string
	OrganizationIban string
	OrganizationBic  string
	CreditTransfers  []TransactionBulkTransferInput
}

// TransactionBulkTransferInput contains all the inputs transfer require executing TransactionBulk use case.
type TransactionBulkTransferInput struct {
	Amount           float64
	Currency         string
	CounterpartyName string
	CounterpartyBic  string
	CounterpartyIban string
	Description      string
}

// AccountBalanceChecker is a storage interface that defines the functionality to check the account balance.
type AccountBalanceChecker interface {
	// BalanceCheck checks whether the account has enough balance or not from a storage.
	//
	// Returns the bank account detail when ever the account has enough balance, otherwise error
	BalanceCheck(ctx context.Context, accountState model.BankAccountState, amount model.Cents) (*model.BankAccount, error)
}

// BalanceUpdater is a storage interface that defines the functionality to update the account balance.
type BalanceUpdater interface {
	// BalanceUpdate updates the account balance from a storage.
	BalanceUpdate(ctx context.Context, accountID model.BankAccountID, amount model.Cents) error
}

// TransactionAdder is a storage interface that defines the functionality to add the transaction.
type TransactionAdder interface {
	// Add adds a transaction into a storage.
	Add(ctx context.Context, transactionStates []model.TransactionState) error
}

type transactionBulk struct {
	logger  ctxd.Logger
	storage *sqluct.Storage
	checker AccountBalanceChecker
	updater BalanceUpdater
	adder   TransactionAdder
}

var _ TransactionBulk = new(transactionBulk)

// NewTransactionBulk creates an instance of TransactionBulk use case.
func NewTransactionBulk(
	logger ctxd.Logger,
	storage *sqluct.Storage,
	checker AccountBalanceChecker,
	didacticer BalanceUpdater,
	adder TransactionAdder,
) TransactionBulk {
	return &transactionBulk{
		logger:  logger,
		storage: storage,
		checker: checker,
		updater: didacticer,
		adder:   adder,
	}
}

// TransactionBulk use case functionality to process a transaction in bulk.
func (tb *transactionBulk) TransactionBulk(ctx context.Context, input TransactionBulkInput) error {
	ctx = ctxd.AddFields(
		ctx,
		"organization_name", input.OrganizationName,
		"organization_bic", input.OrganizationBic,
		"organization_iban", input.OrganizationIban,
	)

	var creditsTransferAmount model.Cents

	for _, transfer := range input.CreditTransfers {
		// convert transfer amount into cents
		creditsTransferAmount += model.ToCents(transfer.Amount)
	}

	ctx = ctxd.AddFields(ctx, "credit_transfers_total", creditsTransferAmount)

	err := tb.storage.InTx(ctx, func(ctx context.Context) error {
		accountState := model.BankAccountState{
			OrganizationName: input.OrganizationName,
			Iban:             input.OrganizationIban,
			Bic:              input.OrganizationBic,
		}

		tb.logger.Debug(ctx, "checking account has enough balance")

		account, err := tb.checker.BalanceCheck(ctx, accountState, creditsTransferAmount)
		if err != nil {
			return err
		}

		tb.logger.Debug(ctx, "account has enough balance")

		var TransactionStates []model.TransactionState

		for _, transfer := range input.CreditTransfers {
			amountCents := model.ToCents(transfer.Amount)

			tb.logger.Debug(ctx, "adding transfer",
				"bankAccount_id", account.ID,
				"transfer_amount", account.ID,
				"counterparty_iban", transfer.CounterpartyIban,
				"counterparty_bic", transfer.CounterpartyBic,
				"amount_cents", amountCents,
			)

			TransactionStates = append(TransactionStates, model.TransactionState{
				CounterpartyName: transfer.CounterpartyName,
				CounterpartyIban: transfer.CounterpartyIban,
				CounterpartyBic:  transfer.CounterpartyBic,
				AmountCents:      amountCents,
				AmountCurrency:   transfer.Currency,
				BankAccountID:    account.ID,
				Description:      transfer.Description,
			})
		}

		err = tb.adder.Add(ctx, TransactionStates)
		if err != nil {
			return err
		}

		tb.logger.Debug(ctx, "all transfer added")

		tb.logger.Debug(ctx, "didactic balance from account")

		newAmount := account.BalanceCents - creditsTransferAmount

		err = tb.updater.BalanceUpdate(ctx, account.ID, newAmount)
		if err != nil {
			return err
		}

		tb.logger.Debug(ctx, "balance from account didactic")

		return nil
	})

	return err
}
