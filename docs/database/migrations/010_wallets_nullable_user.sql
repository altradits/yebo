-- Chama wallets are not owned by a single user; make user_id nullable.
-- Replace the plain UNIQUE constraint with a partial unique index so
-- NULL rows (group wallets) are not subject to uniqueness.
ALTER TABLE wallets ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE wallets DROP CONSTRAINT wallets_user_id_key;
CREATE UNIQUE INDEX wallets_user_id_key ON wallets(user_id) WHERE user_id IS NOT NULL;
