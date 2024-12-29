-- Create temporary tables with UUID columns
CREATE TABLE users_new (
    id BINARY(16) PRIMARY KEY,
    username VARCHAR(30) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(10) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE balances_new (
    user_id BINARY(16) PRIMARY KEY,
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    last_updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users_new(id)
);

CREATE TABLE transactions_new (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_user_id BINARY(16) NOT NULL,
    to_user_id BINARY(16) NOT NULL,
    amount DECIMAL(20,2) NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    notes TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users_new(id),
    FOREIGN KEY (to_user_id) REFERENCES users_new(id)
);

CREATE TABLE balance_history_new (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BINARY(16) NOT NULL,
    old_amount DECIMAL(20,2) NOT NULL,
    new_amount DECIMAL(20,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users_new(id)
);

CREATE TABLE audit_logs_new (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(36) NOT NULL,
    action VARCHAR(50) NOT NULL,
    details TEXT,
    user_id BINARY(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users_new(id)
);

-- Copy data from old tables to new tables with UUID conversion
INSERT INTO users_new (id, username, email, password_hash, role, created_at, updated_at, deleted_at)
SELECT UUID_TO_BIN(UUID()), username, email, password_hash, role, created_at, updated_at, deleted_at
FROM users;

-- Create a temporary table to store the ID mappings
CREATE TABLE id_mappings (
    old_id INTEGER PRIMARY KEY,
    new_id BINARY(16) NOT NULL
);

INSERT INTO id_mappings (old_id, new_id)
SELECT u.id, un.id
FROM users u
JOIN users_new un ON u.username = un.username;

-- Copy data to other tables using the ID mappings
INSERT INTO balances_new (user_id, amount, last_updated_at, created_at, updated_at, deleted_at)
SELECT m.new_id, b.amount, b.last_updated_at, b.created_at, b.updated_at, b.deleted_at
FROM balances b
JOIN id_mappings m ON b.user_id = m.old_id;

INSERT INTO transactions_new (id, from_user_id, to_user_id, amount, type, status, notes, created_at, updated_at, deleted_at)
SELECT t.id, mf.new_id, mt.new_id, t.amount, t.type, t.status, t.notes, t.created_at, t.updated_at, t.deleted_at
FROM transactions t
JOIN id_mappings mf ON t.from_user_id = mf.old_id
JOIN id_mappings mt ON t.to_user_id = mt.old_id;

INSERT INTO balance_history_new (id, user_id, old_amount, new_amount, created_at, deleted_at)
SELECT bh.id, m.new_id, bh.old_amount, bh.new_amount, bh.created_at, bh.deleted_at
FROM balance_history bh
JOIN id_mappings m ON bh.user_id = m.old_id;

INSERT INTO audit_logs_new (id, entity_type, entity_id, action, details, user_id, created_at, updated_at, deleted_at)
SELECT a.id, a.entity_type,
    CASE
        WHEN a.entity_type = 'user' THEN BIN_TO_UUID(m2.new_id)
        ELSE a.entity_id
    END,
    a.action, a.details, m.new_id, a.created_at, a.updated_at, a.deleted_at
FROM audit_logs a
JOIN id_mappings m ON a.user_id = m.old_id
LEFT JOIN id_mappings m2 ON a.entity_type = 'user' AND a.entity_id = m2.old_id;

-- Drop old tables and rename new tables
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS balance_history;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS balances;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS id_mappings;

ALTER TABLE users_new RENAME TO users;
ALTER TABLE balances_new RENAME TO balances;
ALTER TABLE transactions_new RENAME TO transactions;
ALTER TABLE balance_history_new RENAME TO balance_history;
ALTER TABLE audit_logs_new RENAME TO audit_logs;

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