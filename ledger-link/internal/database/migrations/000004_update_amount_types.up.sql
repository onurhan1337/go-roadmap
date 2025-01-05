-- Update balances table
ALTER TABLE balances MODIFY COLUMN amount DECIMAL(20,8) NOT NULL DEFAULT 0;

-- Update transactions table
ALTER TABLE transactions MODIFY COLUMN amount DECIMAL(20,8) NOT NULL;

-- Update balance_history table
ALTER TABLE balance_history MODIFY COLUMN old_amount DECIMAL(20,8) NOT NULL;
ALTER TABLE balance_history MODIFY COLUMN new_amount DECIMAL(20,8) NOT NULL; 