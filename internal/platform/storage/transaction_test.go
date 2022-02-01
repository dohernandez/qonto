package storage_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
	"github.com/dohernandez/qonto/internal/platform/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction_Add(t *testing.T) {
	t.Parallel()

	type args struct {
		transactionState []model.TransactionState
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		pgxErr  error
		err     error
	}{
		{
			name: "insert transaction successfully",
			args: args{
				transactionState: []model.TransactionState{
					{
						CounterpartyName: "CounterpartyName1",
						CounterpartyIban: "CounterpartyIban1",
						CounterpartyBic:  "CounterpartyBic1",
						AmountCents:      1000,
						AmountCurrency:   "EUR",
						BankAccountID:    1,
						Description:      "Description1",
					},
					{
						CounterpartyName: "CounterpartyName2",
						CounterpartyIban: "CounterpartyIban2",
						CounterpartyBic:  "CounterpartyBic2",
						AmountCents:      2000,
						AmountCurrency:   "EUR",
						BankAccountID:    1,
						Description:      "Description2",
					},
				},
			},
			wantErr: false,
			pgxErr:  nil,
			err:     nil,
		},
		{
			name: "insert transaction fail",
			args: args{
				transactionState: []model.TransactionState{
					{
						CounterpartyName: "CounterpartyName",
						CounterpartyIban: "CounterpartyIban",
						CounterpartyBic:  "CounterpartyBic",
						AmountCents:      1000,
						AmountCurrency:   "EUR",
						BankAccountID:    1,
						Description:      "Description",
					},
				},
			},
			wantErr: true,
			pgxErr:  sql.ErrTxDone,
			err:     ctxd.WrapError(context.Background(), sql.ErrTxDone, "storage.Transaction: failed to add transaction"),
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			var (
				values   string
				withArgs []driver.Value
				i        int
			)

			for _, state := range tc.args.transactionState {
				if values != "" {
					values += ","
				}

				values += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d)", i+1, i+2, i+3, i+4, i+5, i+6, i+7)

				withArgs = append(withArgs,
					state.CounterpartyName,
					state.CounterpartyIban,
					state.CounterpartyBic,
					state.AmountCents,
					state.AmountCurrency,
					state.BankAccountID,
					state.Description,
				)

				i += 7
			}

			meQuery := mock.ExpectExec(`
				INSERT INTO transactions (counterparty_name,counterparty_iban,counterparty_bic,amount_cents,amount_currency,bank_account_id,description) 
				VALUES ` + values + `
			`).WithArgs(withArgs...)

			if tc.err == nil {
				meQuery.WillReturnResult(sqlmock.NewResult(2, 2))
			} else {
				meQuery.WillReturnError(tc.pgxErr)
			}

			st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

			r := storage.NewTransaction(st)

			if err = r.Add(context.Background(), tc.args.transactionState); (err != nil) != tc.wantErr {
				t.Errorf("Add() error = %v, wantErr %v", err, tc.wantErr)
			}

			assert.ErrorIsf(t, tc.err, err, "Add() err got = %v, want %v", err, tc.err)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Add() expectations were not met = %v", err)
			}
		})
	}
}
