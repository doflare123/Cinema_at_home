ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_id BIGINT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS telegram_username VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS status VARCHAR(16);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_status_check'
    ) THEN
        ALTER TABLE users
            ADD CONSTRAINT users_status_check CHECK (status IN ('pending', 'active', 'rejected', 'blocked'));
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_index idx
        JOIN pg_class tbl ON tbl.oid = idx.indrelid
        JOIN pg_attribute attr ON attr.attrelid = tbl.oid
        WHERE tbl.relname = 'users'
          AND idx.indisunique
          AND attr.attnum = ANY(idx.indkey)
          AND attr.attname = 'telegram_id'
          AND array_length(idx.indkey, 1) = 1
    ) THEN
        CREATE UNIQUE INDEX idx_users_telegram_id ON users(telegram_id);
    END IF;
END $$;

UPDATE users SET display_name = username WHERE display_name IS NULL;
UPDATE users SET status = 'active' WHERE status IS NULL;

ALTER TABLE users ALTER COLUMN display_name SET NOT NULL;
ALTER TABLE users ALTER COLUMN status SET DEFAULT 'pending';
ALTER TABLE users ALTER COLUMN status SET NOT NULL;
