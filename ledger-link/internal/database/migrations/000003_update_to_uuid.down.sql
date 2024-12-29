-- Create temporary tables with integer IDs
CREATE TABLE users_old (
    id SERIAL PRIMARY KEY,
    username VARCHAR(30) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE balances_old (
    user_id INTEGER PRIMARY KEY REFERENCES users_old(id),
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE transactions_old (
    id SERIAL PRIMARY KEY,
    from_user_id INTEGER NOT NULL REFERENCES users_old(id),
    to_user_id INTEGER NOT NULL REFERENCES users_old(id),
    amount DECIMAL(20,2) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE balance_history_old (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users_old(id),
    old_amount DECIMAL(20,2) NOT NULL,
    new_amount DECIMAL(20,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE audit_logs_old (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id TEXT NOT NULL,
    action VARCHAR(50) NOT NULL,
    details TEXT,
    user_id INTEGER NOT NULL REFERENCES users_old(id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Copy data from UUID tables to integer ID tables
INSERT INTO users_old (username, email, password_hash, role, created_at, updated_at, deleted_at)
SELECT username, email, password_hash, role, created_at, updated_at, deleted_at
FROM users;

-- Create a temporary table to store the ID mappings
CREATE TABLE id_mappings (
    old_uuid UUID PRIMARY KEY,
    new_id INTEGER NOT NULL
);

INSERT INTO id_mappings (old_uuid, new_id)
SELECT u.id, uo.id
FROM users u
JOIN users_old uo ON u.username = uo.username;

-- Copy data to other tables using the ID mappings
INSERT INTO balances_old (user_id, amount, last_updated_at, created_at, updated_at, deleted_at)
SELECT m.new_id, b.amount, b.last_updated_at, b.created_at, b.updated_at, b.deleted_at
FROM balances b
JOIN id_mappings m ON b.user_id = m.old_uuid;

INSERT INTO transactions_old (id, from_user_id, to_user_id, amount, type, status, notes, created_at, updated_at, deleted_at)
SELECT t.id, mf.new_id, mt.new_id, t.amount, t.type, t.status, t.notes, t.created_at, t.updated_at, t.deleted_at
FROM transactions t
JOIN id_mappings mf ON t.from_user_id = mf.old_uuid
JOIN id_mappings mt ON t.to_user_id = mf.old_uuid;

INSERT INTO balance_history_old (id, user_id, old_amount, new_amount, created_at, deleted_at)
SELECT bh.id, m.new_id, bh.old_amount, bh.new_amount, bh.created_at, bh.deleted_at
FROM balance_history bh
JOIN id_mappings m ON bh.user_id = m.old_uuid;

INSERT INTO audit_logs_old (id, entity_type, entity_id, action, details, user_id, created_at, updated_at, deleted_at)
SELECT a.id, a.entity_type,
    CASE
        WHEN a.entity_type = 'user' THEN m2.new_id::text
        ELSE a.entity_id
    END,
    a.action, a.details, m.new_id, a.created_at, a.updated_at, a.deleted_at
FROM audit_logs a
JOIN id_mappings m ON a.user_id = m.old_uuid
LEFT JOIN id_mappings m2 ON a.entity_type = 'user' AND a.entity_id = m2.old_uuid::text;

-- Drop UUID tables and rename integer ID tables
DROP TABLE audit_logs;
DROP TABLE balance_history;
DROP TABLE transactions;
DROP TABLE balances;
DROP TABLE users;
DROP TABLE id_mappings;

ALTER TABLE users_old RENAME TO users;
ALTER TABLE balances_old RENAME TO balances;
ALTER TABLE transactions_old RENAME TO transactions;
ALTER TABLE balance_history_old RENAME TO balance_history;
ALTER TABLE audit_logs_old RENAME TO audit_logs;

-- Create indexes
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_balances_deleted_at ON balances(deleted_at);
CREATE INDEX idx_transactions_deleted_at ON transactions(deleted_at);
CREATE INDEX idx_transactions_from_user_id ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user_id ON transactions(to_user_id);
CREATE INDEX idx_balance_history_user_id ON balance_history(user_id);
CREATE INDEX idx_audit_logs_entity_type_entity_id ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_deleted_at ON audit_logs(deleted_at);