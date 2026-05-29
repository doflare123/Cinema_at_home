CREATE TABLE IF NOT EXISTS import_runs (
    id SERIAL PRIMARY KEY,
    source_file_name TEXT NOT NULL,
    imported_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    dry_run BOOLEAN NOT NULL DEFAULT FALSE,
    rows_total INT NOT NULL DEFAULT 0,
    rows_parsed INT NOT NULL DEFAULT 0,
    rows_created INT NOT NULL DEFAULT 0,
    rows_skipped_existing INT NOT NULL DEFAULT 0,
    rows_skipped_invalid INT NOT NULL DEFAULT 0,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_import_runs_created_at
    ON import_runs(created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_import_runs_imported_by_created
    ON import_runs(imported_by_user_id, created_at DESC, id DESC);
