Feature: Performing transfers
  As a customers, I want tp perform transfers perform transfers in bulk due to
  performing those transfers one by one would be painful and time consuming.

  Background:
    Given there is a clean "postgres" database
    And these rows are stored in table "bank_accounts" of database "postgres":
      | id | organization_name | balance_cents | iban                        | bic         |
      | 1  | ACME Corp         | 10000000      | FR10474608000002006107XXXXX | OIVUSCLQXXX |

  Scenario: Performing transfers successfully
    When I request HTTP endpoint with method "POST" and URI "/v1/transfer/bulk"
    And I request HTTP endpoint with body from file
    """
    ./features/_testdata/sample1.json
    """
    And I concurrently request idempotent HTTP endpoint

    Then I should have response with status "Created"
    And I should have other responses with status "Unprocessable Entity"
    And these rows are available in table "transactions" of database "postgres":
      | counterparty_name | counterparty_iban           | counterparty_bic | amount_cents | amount_currency | bank_account_id | description                         |
      | Bip Bip           | EE383680981021245685        | CRLYFRPPTOU      | 1450         | EUR             | 1               | Wonderland/4410                     |
      | Wile E Coyote     | DE9935420810036209081725212 | ZDRPLBQI         | 6123800      | EUR             | 1               | //TeslaMotors/Invoice/12            |
      | Bugs Bunny        | FR0010009380540930414023042 | RNJZNTMC         | 99900        | EUR             | 1               | 2020 09 24/2020 09 25/GoldenCarrot/ |
    And these rows are available in table "bank_accounts" of database "postgres":
      | id | balance_cents |
      | 1  | 3774850       |

  Scenario: Unprocessable transfers, not enough balance
    When I request HTTP endpoint with method "POST" and URI "/v1/transfer/bulk"
    And I request HTTP endpoint with body from file
    """
    ./features/_testdata/sample2.json
    """

    Then I should have response with status "Unprocessable Entity"
    And no rows are available in table "transactions" of database "postgres"
    And these rows are available in table "bank_accounts" of database "postgres":
      | id | balance_cents |
      | 1  | 10000000      |