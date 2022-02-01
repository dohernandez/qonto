package model

// TransactionID is the type of Transaction id.
type TransactionID int64

// Transaction represent a transaction.
type Transaction struct {
	ID TransactionID `db:"id"`

	TransactionState
}

// TransactionState represents the Transaction internal state/data.
type TransactionState struct {
	CounterpartyName string        `db:"counterparty_name"`
	CounterpartyIban string        `db:"counterparty_iban"`
	CounterpartyBic  string        `db:"counterparty_bic"`
	AmountCents      Cents         `db:"amount_cents"`
	AmountCurrency   string        `db:"amount_currency"`
	BankAccountID    BankAccountID `db:"bank_account_id"`
	Description      string        `db:"description"`
}
