package model

// BankAccountID is the type of BankAccount id.
type BankAccountID int64

// BankAccount represent a bank account.
type BankAccount struct {
	ID BankAccountID `db:"id"`

	BankAccountState
}

// BankAccountState represents the BankAccount internal state/data.
type BankAccountState struct {
	OrganizationName string `db:"organization_name" json:"organization_name"`
	BalanceCents     Cents  `db:"balance_cents"`
	Iban             string `db:"iban"`
	Bic              string `db:"bic"`
}
