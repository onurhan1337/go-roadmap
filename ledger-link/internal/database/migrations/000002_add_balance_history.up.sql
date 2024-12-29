CREATE TABLE balance_history (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    old_amount DECIMAL(20, 8) NOT NULL,
    new_amount DECIMAL(20, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    KEY idx_balance_history_user_id (user_id),
    KEY idx_balance_history_created_at (created_at),
    KEY idx_balance_history_deleted_at (deleted_at),
    FOREIGN KEY (user_id) REFERENCES users(id)
);