CREATE TABLE users (
    rowid INTEGER PRIMARY KEY,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    email TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL DEFAULT '',
    timezone TEXT NOT NULL DEFAULT 'Australia/Melbourne'
);

CREATE TABLE media_category (
    category TEXT PRIMARY KEY
) WITHOUT ROWID;

CREATE TABLE user_settings (
    user_id INTEGER NOT NULL PRIMARY KEY,
    max_afk INTEGER NOT NULL DEFAULT 60 CHECK (max_afk >= 30 AND max_afk <= 1200),
    max_blur INTEGER NOT NULL DEFAULT 2 CHECK (max_blur >= 0 AND max_blur <=10),
    inactivity_blur INTEGER NOT NULL DEFAULT 2 CHECK (inactivity_blur >= 0 AND inactivity_blur <=10),
    menu_blur INTEGER NOT NULL DEFAULT 8 CHECK (menu_blur >= 0 AND menu_blur <=10),
    last_update TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(rowid)
) WITHOUT ROWID;

CREATE INDEX user_settings_index ON user_settings(user_id, last_update);

CREATE TABLE user_category_settings (
    user_id INTEGER NOT NULL,
    category TEXT NOT NULL,
    max_afk INTEGER NOT NULL DEFAULT 60 CHECK (max_afk >= 30 AND max_afk <= 1200),
    inactivity_blur INTEGER NOT NULL DEFAULT 2 CHECK (inactivity_blur >= 0 AND inactivity_blur <=10),
    menu_blur INTEGER NOT NULL DEFAULT 8 CHECK (menu_blur >= 0 AND menu_blur <=10),
    last_update TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(rowid),
    PRIMARY KEY (category, user_id)
) WITHOUT ROWID;

CREATE INDEX user_category_settings_index ON user_category_settings(user_id, last_update);

CREATE TABLE media (
    identifier TEXT NOT NULL,
    category TEXT NOT NULL,
    series TEXT NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    last_update TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category) REFERENCES media_category(category),
    FOREIGN KEY (user_id) REFERENCES users(rowid),
    PRIMARY KEY (identifier, category, series, user_id)
) WITHOUT ROWID;

CREATE TABLE media_stats (
    media_identifier TEXT NOT NULL,
    category TEXT NOT NULL,
    series TEXT NOT NULL DEFAULT '',
    user_id INTEGER NOT NULL,
    immerse_date TEXT NOT NULL,
    last_read TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    read_time INTEGER NOT NULL DEFAULT 0 CHECK (read_time >= 0),
    read_chars INTEGER NOT NULL DEFAULT 0 CHECK (read_chars >= 0),
    read_lines INTEGER DEFAULT 0 CHECK (read_lines >= 0),
    read_pages INTEGER DEFAULT 0 CHECK (read_pages >= 0),
    paused INTEGER DEFAULT 0 CHECK (paused IN (0, 1)),
    last_update TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_identifier, category, series, user_id) REFERENCES media(identifier, category, series, user_id),
    FOREIGN KEY (category) REFERENCES media_category(category),
    FOREIGN KEY (user_id) REFERENCES users(rowid),
    PRIMARY KEY (media_identifier, category, user_id, immerse_date)
) WITHOUT ROWID;

CREATE INDEX media_stats_user_index ON media_stats(user_id, last_update);
CREATE INDEX media_stats_user_category_index ON media_stats(user_id, category, last_update);
