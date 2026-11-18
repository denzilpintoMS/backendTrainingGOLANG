-- Run with psql after connecting to bank_db.
-- Temporary practice tables keep the API's data untouched.
-- All sample changes are rolled back at the end.
BEGIN;

CREATE TEMP TABLE practice_users (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL
);

-- UNIQUE user_id means each user can have at most one account.
CREATE TEMP TABLE practice_accounts (
    id BIGINT PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL REFERENCES practice_users(id),
    account_number TEXT UNIQUE NOT NULL,
    balance BIGINT NOT NULL CHECK (balance >= 0)
);

-- One account can appear in many incoming and outgoing transactions.
CREATE TEMP TABLE practice_transactions (
    id BIGINT PRIMARY KEY,
    type TEXT NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    from_account_id BIGINT REFERENCES practice_accounts(id),
    to_account_id BIGINT REFERENCES practice_accounts(id)
);

-- INSERT: create sample users, accounts, and money movements (amounts in paise).
INSERT INTO practice_users (id, name, email) VALUES
    (1, 'Alice', 'alice@example.com'),
    (2, 'Bob', 'bob@example.com'),
    (3, 'Charlie', 'charlie@example.com');

INSERT INTO practice_accounts (id, user_id, account_number, balance) VALUES
    (1, 1, 'ACC001', 7500),
    (2, 2, 'ACC002', 7500);

INSERT INTO practice_transactions (id, type, amount, from_account_id, to_account_id) VALUES
    (1, 'deposit', 10000, NULL, 1),
    (2, 'deposit', 5000, NULL, 2),
    (3, 'transfer', 2500, 1, 2);

-- SELECT + WHERE: find a specific account.
SELECT id, account_number, balance
FROM practice_accounts
WHERE user_id = 1;

-- ORDER BY: show newest transaction IDs first.
SELECT id, type, amount
FROM practice_transactions
ORDER BY id DESC;

-- GROUP BY: count and total the amounts for each transaction type.
SELECT type, COUNT(*) AS transaction_count, SUM(amount) AS total_amount
FROM practice_transactions
GROUP BY type
ORDER BY type;

-- INNER JOIN: Alice and Bob appear because they have matching accounts.
SELECT u.name, a.account_number, a.balance
FROM practice_users u
INNER JOIN practice_accounts a ON a.user_id = u.id
ORDER BY u.id;

-- LEFT JOIN: Charlie also appears, with NULL account data.
SELECT u.name, a.account_number, a.balance
FROM practice_users u
LEFT JOIN practice_accounts a ON a.user_id = u.id
ORDER BY u.id;

-- One-to-many: show each account's incoming transaction records.
SELECT a.account_number, t.id AS transaction_id, t.type, t.amount
FROM practice_accounts a
INNER JOIN practice_transactions t ON t.to_account_id = a.id
ORDER BY a.id, t.id;

-- UPDATE: change a user's name and read the result.
UPDATE practice_users SET name = 'Alice Updated' WHERE id = 1;
SELECT id, name FROM practice_users WHERE id = 1;

-- DELETE: remove the sample user who has no account.
DELETE FROM practice_users WHERE id = 3;
SELECT id, name FROM practice_users ORDER BY id;

-- Undo all sample changes, including the temporary table definitions.
ROLLBACK;
