create table bank_accounts
(
    id                serial primary key,
    organization_name TEXT,
    balance_cents     INTEGER,
    iban              TEXT,
    bic               TEXT
);

INSERT INTO bank_accounts (id, organization_name, balance_cents, iban, bic)
VALUES (1, 'ACME Corp', 10000000, 'FR10474608000002006107XXXXX', 'OIVUSCLQXXX');

create table transactions
(
    id                serial primary key,
    counterparty_name TEXT,
    counterparty_iban TEXT,
    counterparty_bic  TEXT,
    amount_cents      INTEGER,
    amount_currency   TEXT,
    bank_account_id   INTEGER,
    description       TEXT,

    FOREIGN KEY (bank_account_id) REFERENCES bank_accounts (id)
);

INSERT INTO transactions (counterparty_name, counterparty_iban, counterparty_bic, amount_cents, amount_currency,
                          bank_account_id, description)
VALUES ('ACME Corp. Main Account', 'EE382200221020145685', 'CCOPFRPPXXX', 11000000, 'EUR', 1, 'Treasury income'),
('Bip Bip', 'EE383680981021245685', 'CRLYFRPPTOU', -1000000, 'EUR', 1, 'Bip Bip Salary');