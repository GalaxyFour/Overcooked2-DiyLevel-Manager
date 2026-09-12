PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_md5 TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL CHECK(role IN ('author', 'admin', 'super_admin')),
    must_change_password INTEGER NOT NULL DEFAULT 1,
    display_name TEXT NOT NULL DEFAULT '',
    avatar_cos_key TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    disabled_at TEXT
);

CREATE TABLE IF NOT EXISTS level_sets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    author_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL DEFAULT '',
    name_zh TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'hidden', 'invalid')),
    cover_cos_key TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_level_sets_author ON level_sets(author_id);
CREATE INDEX IF NOT EXISTS idx_level_sets_status ON level_sets(status);

CREATE TABLE IF NOT EXISTS level_set_versions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    set_id INTEGER NOT NULL REFERENCES level_sets(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    uid TEXT NOT NULL DEFAULT '',
    zip_cos_key TEXT NOT NULL DEFAULT '',
    file_count INTEGER NOT NULL DEFAULT 0,
    has_runtime INTEGER NOT NULL DEFAULT 0,
    has_common_w2 INTEGER NOT NULL DEFAULT 0,
    parse_status TEXT NOT NULL DEFAULT 'queued',
    parse_phase TEXT NOT NULL DEFAULT '',
    parse_progress INTEGER NOT NULL DEFAULT 0,
    parse_message TEXT NOT NULL DEFAULT '',
    is_latest INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(set_id, version)
);

CREATE INDEX IF NOT EXISTS idx_versions_set ON level_set_versions(set_id);
CREATE INDEX IF NOT EXISTS idx_versions_latest ON level_set_versions(set_id, is_latest);

CREATE TABLE IF NOT EXISTS level_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL REFERENCES level_set_versions(id) ON DELETE CASCADE,
    level_id TEXT NOT NULL,
    level_name TEXT NOT NULL DEFAULT '',
    level_name_zh TEXT NOT NULL DEFAULT '',
    scene_name TEXT NOT NULL DEFAULT '',
    screenshot_cos_key TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_entries_version ON level_entries(version_id);

CREATE TABLE IF NOT EXISTS parse_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    version_id INTEGER NOT NULL REFERENCES level_set_versions(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    temp_dir TEXT NOT NULL DEFAULT '',
    started_at TEXT,
    finished_at TEXT,
    error TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_parse_jobs_version ON parse_jobs(version_id);
CREATE INDEX IF NOT EXISTS idx_parse_jobs_status ON parse_jobs(status);

CREATE TABLE IF NOT EXISTS presign_cache (
    cos_key TEXT PRIMARY KEY,
    url TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    kind TEXT NOT NULL CHECK(kind IN ('package', 'image'))
);
