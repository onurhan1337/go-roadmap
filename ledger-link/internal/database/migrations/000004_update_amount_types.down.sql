-- Revert balances table
ALTER TABLE balances MODIFY COLUMN amount DOUBLE NOT NULL DEFAULT 0;

-- Revert transactions table
ALTER TABLE transactions MODIFY COLUMN amount DOUBLE NOT NULL;

-- Revert balance_history table
ALTER TABLE balance_history MODIFY COLUMN old_amount DOUBLE NOT NULL;
ALTER TABLE balance_history MODIFY COLUMN new_amount DOUBLE NOT NULL; 