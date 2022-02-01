package storage_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/dohernandez/qonto/internal/domain/model"
	"github.com/dohernandez/qonto/internal/platform/storage"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRowsClosed = errors.New("sql: Rows are closed")

func TestBankAccount_BalanceCheck(t *testing.T) {
	t.Parallel()

	type args struct {
		amount       model.Cents
		accountState model.BankAccountState
		pgxResult    *model.BankAccount
	}

	tests := []struct {
		name    string
		args    args
		want    *model.BankAccount
		wantErr bool
		pgxErr  error
		err     error
	}{
		{
			name: "account has enough balance",
			args: args{
				amount: 10000,
				accountState: model.BankAccountState{
					OrganizationName: "OrganizationName",
					BalanceCents:     0,
					Iban:             "Iban",
					Bic:              "Bic",
				},
				pgxResult: &model.BankAccount{
					ID: 1,
					BankAccountState: model.BankAccountState{
						OrganizationName: "OrganizationName",
						BalanceCents:     1000000,
						Iban:             "Iban",
						Bic:              "Bic",
					},
				},
			},
			want: &model.BankAccount{
				ID: 1,
				BankAccountState: model.BankAccountState{
					OrganizationName: "OrganizationName",
					BalanceCents:     1000000,
					Iban:             "Iban",
					Bic:              "Bic",
				},
			},
			wantErr: false,
			pgxErr:  nil,
			err:     nil,
		},
		{
			name: "account does not exists",
			args: args{
				amount: 10000,
				accountState: model.BankAccountState{
					OrganizationName: "Organization",
					BalanceCents:     0,
					Iban:             "Iban",
					Bic:              "Bic",
				},
				pgxResult: nil,
			},
			want:    nil,
			wantErr: true,
			pgxErr:  sql.ErrNoRows,
			err:     ctxd.WrapError(context.Background(), storage.ErrNotFound, "storage.BankAccount: failed to check account balance"),
		},
		{
			name: "not enough balance",
			args: args{
				amount: 100000,
				accountState: model.BankAccountState{
					OrganizationName: "Organization",
					BalanceCents:     0,
					Iban:             "Iban",
					Bic:              "Bic",
				},
				pgxResult: &model.BankAccount{
					ID: 1,
					BankAccountState: model.BankAccountState{
						OrganizationName: "OrganizationName",
						BalanceCents:     1000,
						Iban:             "Iban",
						Bic:              "Bic",
					},
				},
			},
			want:    nil,
			wantErr: true,
			pgxErr:  nil,
			err:     ctxd.WrapError(context.Background(), storage.ErrNotEnoughBalance, "storage.BankAccount: failed to check account balance"),
		},
		{
			name: "db error when account balance check",
			args: args{
				amount: 100000,
				accountState: model.BankAccountState{
					OrganizationName: "Organization",
					BalanceCents:     0,
					Iban:             "Iban",
					Bic:              "Bic",
				},
				pgxResult: nil,
			},
			want:    nil,
			wantErr: true,
			pgxErr:  errRowsClosed,
			err:     ctxd.WrapError(context.Background(), errRowsClosed, "storage.BankAccount: failed to check account balance"),
		},
	}
	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			meQuery := mock.ExpectQuery(`
				SELECT id, organization_name, balance_cents, iban, bic 
				FROM bank_accounts  
				WHERE organization_name = $1 AND iban = $2 AND bic = $3
			`).
				WithArgs(tc.args.accountState.OrganizationName, tc.args.accountState.Iban, tc.args.accountState.Bic)

			if tc.args.pgxResult != nil {
				rows := sqlmock.NewRows([]string{
					"id", "organization_name", "balance_cents", "iban", "bic",
				})

				rows.AddRow(
					tc.args.pgxResult.ID, tc.args.pgxResult.OrganizationName, tc.args.pgxResult.BalanceCents, tc.args.pgxResult.Iban, tc.args.pgxResult.Bic,
				)

				meQuery.WillReturnRows(rows)
			} else {
				meQuery.WillReturnError(tc.pgxErr)
			}

			st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

			r := storage.NewBankAccount(st)

			got, err := r.BalanceCheck(context.Background(), tc.args.accountState, tc.args.amount)
			if (err != nil) != tc.wantErr {
				t.Errorf("BalanceCheck() error = %v, wantErr %v", err, tc.wantErr)

				return
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("BalanceCheck() got = %v, want %v", got, tc.want)
			}

			assert.ErrorIsf(t, tc.err, err, "BalanceCheck() err got = %v, want %v", err, tc.err)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("BalanceCheck() expectations were not met = %v", err)
			}
		})
	}
}

func TestBankAccount_BalanceUpdate(t *testing.T) {
	t.Parallel()

	type args struct {
		accountID model.BankAccountID
		amount    model.Cents
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
		pgxErr  error
		err     error
	}{
		{
			name: "balance didactic successfully",
			args: args{
				accountID: 1,
				amount:    10000,
			},
			wantErr: false,
			pgxErr:  nil,
			err:     nil,
		},
		{
			name: "balance didactic successfully",
			args: args{
				accountID: 1,
				amount:    10000,
			},
			wantErr: true,
			pgxErr:  errRowsClosed,
			err:     ctxd.WrapError(context.Background(), errRowsClosed, "storage.BankAccount: failed to update account balance"),
		},
	}

	for _, tt := range tests {
		tc := tt

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			require.NoError(t, err)

			meQuery := mock.ExpectExec(`
				UPDATE bank_accounts  
				SET balance_cents = $1
				WHERE id = $2
			`).
				WithArgs(tc.args.amount, tc.args.accountID)

			if tc.pgxErr == nil {
				meQuery.WillReturnResult(sqlmock.NewResult(0, 1))
			} else {
				meQuery.WillReturnError(tc.pgxErr)
			}

			st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

			r := storage.NewBankAccount(st)

			if err = r.BalanceUpdate(context.Background(), tc.args.accountID, tc.args.amount); (err != nil) != tc.wantErr {
				t.Errorf("BalanceUpdate() error = %v, wantErr %v", err, tc.wantErr)
			}

			assert.ErrorIsf(t, tc.err, err, "BalanceUpdate() err got = %v, want %v", err, tc.err)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("BalanceUpdate() expectations were not met = %v", err)
			}
		})
	}
}

func TestBankAccount_BalanceCheck_ForUpdate(t *testing.T) {
	t.Parallel()

	accountState := model.BankAccountState{
		OrganizationName: "OrganizationName",
		BalanceCents:     0,
		Iban:             "Iban",
		Bic:              "Bic",
	}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	require.NoError(t, err)

	mock.ExpectBegin()

	meQuery := mock.ExpectQuery(`
				SELECT id, organization_name, balance_cents, iban, bic 
				FROM bank_accounts  
				WHERE organization_name = $1 AND iban = $2 AND bic = $3
				FOR UPDATE
			`).
		WithArgs(accountState.OrganizationName, accountState.Iban, accountState.Bic)

	rows := sqlmock.NewRows([]string{
		"id", "organization_name", "balance_cents", "iban", "bic",
	})

	rows.AddRow(
		1, accountState.OrganizationName, 100000, accountState.Iban, accountState.Bic,
	)

	meQuery.WillReturnRows(rows)

	mock.ExpectCommit()

	st := sqluct.NewStorage(sqlx.NewDb(db, "sqlmock"))

	r := storage.NewBankAccount(st)

	err = st.InTx(context.Background(), func(ctx context.Context) error {
		_, err = r.BalanceCheck(ctx, accountState, 10000)

		return err
	})
	assert.NoError(t, err, "BalanceCheck() got error = %v", err)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("BalanceCheck() expectations were not met = %v", err)
	}
}
