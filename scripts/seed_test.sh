#!/usr/bin/env bash
set -euo pipefail
set -a && source .env && set +a

psql "${DB_URL}" <<SQL
-- Test customer
INSERT INTO users (phone, password_hash, full_name, role, kyc_status)
VALUES ('+254711000001', 'test_hash', 'Alice Mwangi', 'customer', 'verified')
ON CONFLICT (phone) DO NOTHING;

INSERT INTO wallets (user_id)
SELECT id FROM users WHERE phone='+254711000001'
ON CONFLICT (user_id) DO NOTHING;

-- Test agent
INSERT INTO users (phone, password_hash, full_name, role)
VALUES ('+254722000002', 'test_hash', 'Bob Kamau', 'agent')
ON CONFLICT (phone) DO NOTHING;

INSERT INTO wallets (user_id)
SELECT id FROM users WHERE phone='+254722000002'
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO agents (user_id, business_name, location, status)
SELECT id, 'Kamau Express', 'Nairobi CBD', 'active'
FROM users WHERE phone='+254722000002'
ON CONFLICT (user_id) DO NOTHING;
SQL

echo "Test data seeded."
