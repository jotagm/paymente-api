CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS accounts (
    id         TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_name TEXT        NOT NULL,
    balance    NUMERIC(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transfers (
    id           TEXT PRIMARY KEY DEFAULT gen_random_uuid(),
    from_account TEXT        NOT NULL REFERENCES accounts(id),
    to_account   TEXT        NOT NULL REFERENCES accounts(id),
    amount       NUMERIC(15,2) NOT NULL,
    status       TEXT        NOT NULL DEFAULT 'completed',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO accounts (id, owner_name, balance) VALUES
    ('acc-001', 'Alice',  10000.00),
    ('acc-002', 'Bob',     5000.00),
    ('acc-003', 'Charlie', 2500.00)
ON CONFLICT DO NOTHING;
