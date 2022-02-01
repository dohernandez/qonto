package usecase_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
	"github.com/dohernandez/qonto/internal/domain/usecase"
	"github.com/dohernandez/qonto/internal/platform/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type accountBalanceCheckerMock struct {
	t            *testing.T
	accountState model.BankAccountState
	amount       model.Cents

	bankAccount *model.BankAccount
	err         error
}

func (abcm *accountBalanceCheckerMock) BalanceCheck(_ context.Context, accountState model.BankAccountState, amount model.Cents) (*model.BankAccount, error) {
	if !reflect.DeepEqual(accountState, abcm.accountState) {
		abcm.t.Errorf("BalanceCheck() got accountState arg = %v, expected %v", accountState, abcm.accountState)
	}

	assert.Equal(abcm.t, abcm.amount, amount, "BalanceCheck() got amount arg = %v, expected %v", amount, abcm.amount)

	return abcm.bankAccount, abcm.err
}

type balanceUpdaterMock struct {
	t *testing.T

	accountID model.BankAccountID
	amount    model.Cents
	err       error
}

func (bdm *balanceUpdaterMock) BalanceUpdate(ctx context.Context, accountID model.BankAccountID, amount model.Cents) error {
	assert.Equal(bdm.t, bdm.accountID, accountID, "BalanceUpdate() got diff accountID arg = %v, expected %v", accountID, bdm.accountID)
	assert.Equal(bdm.t, bdm.amount, amount, "BalanceUpdate() got diff amount arg = %v, expected %v", amount, bdm.amount)

	return bdm.err
}

type transactionAdderMock struct {
	t *testing.T

	transactionStates []model.TransactionState
	err               error
}

func (tam *transactionAdderMock) Add(_ context.Context, transactionStates []model.TransactionState) error {
	assert.Equal(tam.t, tam.transactionStates, transactionStates, "BalanceUpdate() got account arg = %v, expected one of %v", transactionStates, tam.transactionStates)

	return tam.err
}

func Test_transactionBulk_TransactionBulk(t *testing.T) {
	t.Parallel()

	organizationName := "OrganizationName"
	balanceCents := model.Cents(10000)
	iban := "Iban"
	bic := "Bic"

	bankAccount := model.BankAccount{
		ID: model.BankAccountID(1),
		BankAccountState: model.BankAccountState{
			OrganizationName: organizationName,
			BalanceCents:     model.Cents(100000),
			Iban:             iban,
			Bic:              bic,
		},
	}

	transactionStates := []model.TransactionState{
		{
			CounterpartyName: "CounterpartyName1",
			CounterpartyIban: "CounterpartyIban1",
			CounterpartyBic:  "CounterpartyBic1",
			AmountCents:      balanceCents / 2,
			AmountCurrency:   "EUR",
			BankAccountID:    bankAccount.ID,
			Description:      "Description1",
		},
		{
			CounterpartyName: "CounterpartyName1",
			CounterpartyIban: "CounterpartyIban1",
			CounterpartyBic:  "CounterpartyBic1",
			AmountCents:      balanceCents / 2,
			AmountCurrency:   "EUR",
			BankAccountID:    bankAccount.ID,
			Description:      "Description1",
		},
	}

	creditTransfer := []usecase.TransactionBulkTransferInput{
		{
			Amount:           float64(transactionStates[0].AmountCents / 100),
			Currency:         transactionStates[0].AmountCurrency,
			CounterpartyName: transactionStates[0].CounterpartyName,
			CounterpartyBic:  transactionStates[0].CounterpartyBic,
			CounterpartyIban: transactionStates[0].CounterpartyIban,
			Description:      transactionStates[0].Description,
		},
		{
			Amount:           float64(transactionStates[1].AmountCents / 100),
			Currency:         transactionStates[1].AmountCurrency,
			CounterpartyName: transactionStates[1].CounterpartyName,
			CounterpartyBic:  transactionStates[1].CounterpartyBic,
			CounterpartyIban: transactionStates[1].CounterpartyIban,
			Description:      transactionStates[1].Description,
		},
	}

	type fields struct {
		checker usecase.AccountBalanceChecker
		updater usecase.BalanceUpdater
		adder   usecase.TransactionAdder
	}

	type args struct {
		input usecase.TransactionBulkInput
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		err     error
	}{
		{
			name: "transaction proceed successfully",
			fields: fields{
				checker: &accountBalanceCheckerMock{
					t: t,
					accountState: model.BankAccountState{
						OrganizationName: organizationName,
						Iban:             iban,
						Bic:              bic,
					},
					amount:      balanceCents,
					bankAccount: &bankAccount,
					err:         nil,
				},
				updater: &balanceUpdaterMock{
					t:         t,
					accountID: bankAccount.ID,
					amount:    bankAccount.BalanceCents - balanceCents,
					err:       nil,
				},
				adder: &transactionAdderMock{
					t:                 t,
					transactionStates: transactionStates,
				},
			},
			args: args{
				input: usecase.TransactionBulkInput{
					OrganizationName: organizationName,
					OrganizationIban: iban,
					OrganizationBic:  bic,
					CreditTransfers:  creditTransfer,
				},
			},
			wantErr: false,
			err:     nil,
		},
		{
			name: "transaction proceed failed, not enough balance",
			fields: fields{
				checker: &accountBalanceCheckerMock{
					t: t,
					accountState: model.BankAccountState{
						OrganizationName: organizationName,
						Iban:             iban,
						Bic:              bic,
					},
					amount:      balanceCents,
					bankAccount: nil,
					err:         ctxd.WrapError(context.Background(), storage.ErrNotEnoughBalance, "storage.BankAccount: failed to check account balance"),
				},
				updater: nil,
				adder:   nil,
			},
			args: args{
				input: usecase.TransactionBulkInput{
					OrganizationName: organizationName,
					OrganizationIban: iban,
					OrganizationBic:  bic,
					CreditTransfers:  creditTransfer,
				},
			},
			wantErr: true,
			err:     ctxd.WrapError(context.Background(), storage.ErrNotEnoughBalance, "storage.BankAccount: failed to check account balance"),
		},
		{
			name: "transaction proceed failed, error update balance",
			fields: fields{
				checker: &accountBalanceCheckerMock{
					t: t,
					accountState: model.BankAccountState{
						OrganizationName: organizationName,
						Iban:             iban,
						Bic:              bic,
					},
					amount:      balanceCents,
					bankAccount: &bankAccount,
					err:         nil,
				},
				updater: &balanceUpdaterMock{
					t:         t,
					accountID: bankAccount.ID,
					amount:    bankAccount.BalanceCents - balanceCents,
					err:       ctxd.WrapError(context.Background(), sql.ErrTxDone, "storage.BankAccount: failed to update account balance"),
				},
				adder: &transactionAdderMock{
					t:                 t,
					transactionStates: transactionStates,
				},
			},
			args: args{
				input: usecase.TransactionBulkInput{
					OrganizationName: organizationName,
					OrganizationIban: iban,
					OrganizationBic:  bic,
					CreditTransfers:  creditTransfer,
				},
			},
			wantErr: true,
			err:     ctxd.WrapError(context.Background(), sql.ErrTxDone, "storage.BankAccount: failed to update account balance"),
		},
		{
			name: "transaction proceed failed, error adding transfer",
			fields: fields{
				checker: &accountBalanceCheckerMock{
					t: t,
					accountState: model.BankAccountState{
						OrganizationName: organizationName,
						Iban:             iban,
						Bic:              bic,
					},
					amount:      balanceCents,
					bankAccount: &bankAccount,
					err:         nil,
				},
				updater: &balanceUpdaterMock{
					t:         t,
					accountID: bankAccount.ID,
					amount:    bankAccount.BalanceCents - balanceCents,
					err:       nil,
				},
				adder: &transactionAdderMock{
					t:                 t,
					transactionStates: transactionStates,
					err:               ctxd.WrapError(context.Background(), sql.ErrTxDone, "storage.Transaction: failed to add transaction"),
				},
			},
			args: args{
				input: usecase.TransactionBulkInput{
					OrganizationName: organizationName,
					OrganizationIban: iban,
					OrganizationBic:  bic,
					CreditTransfers:  creditTransfer,
				},
			},
			wantErr: true,
			err:     ctxd.WrapError(context.Background(), sql.ErrTxDone, "storage.Transaction: failed to add transaction"),
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			mock.ExpectBegin()

			if !tc.wantErr {
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

			tb := usecase.NewTransactionBulk(ctxd.NoOpLogger{}, st, tc.fields.checker, tc.fields.updater, tc.fields.adder)

			if err = tb.TransactionBulk(context.Background(), tc.args.input); (err != nil) != tc.wantErr {
				t.Errorf("TransactionBulk() error = %v, wantErr %v", err, tc.wantErr)
			}

			assert.ErrorIsf(t, tc.err, err, "TransactionBulk() err got = %v, want %v", err, tc.err)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("TransactionBulk() expectations were not met = %v", err)
			}
		})
	}
}
