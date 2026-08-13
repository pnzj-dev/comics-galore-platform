CREATE UNIQUE INDEX IF NOT EXISTS idx_users_sub_partner_id
ON users(sub_partner_id)
WHERE sub_partner_id IS NOT NULL;
