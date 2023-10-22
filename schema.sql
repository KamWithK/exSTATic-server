CREATE TABLE users (
    rowid INTEGER PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT (datetime('now', 'utc')),
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'Australia/Melbourne'
);
