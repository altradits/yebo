-- Allow chama group wallets to have no owner (user_id = NULL).
-- Replace the plain UNIQUE constraint with a partial unique index
-- so NULL rows are not subject to uniqueness checks.
ALTER TABLE wallets ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE wallets DROP CONSTRAINT wallets_user_id_key;
CREATE UNIQUE INDEX wallets_user_id_key ON wallets(user_id) WHERE user_id IS NOT NULL;
