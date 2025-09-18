-- +goose Up
CREATE TABLE IF NOT EXISTS trade_links
(
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    url         TEXT    NOT NULL,
    description TEXT,
    selected    INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS settings
(
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    key   TEXT NOT NULL UNIQUE,
    value TEXT
);
CREATE TABLE IF NOT EXISTS live_search_settings
(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT    NOT NULL,
    enabled INTEGER NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS trade_links;
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS live_search_settings;
