-- +goose Up
-- +goose StatementBegin
CREATE TABLE logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL,
    module TEXT NOT NULL,
    level TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata TEXT DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better query performance
CREATE INDEX idx_logs_timestamp ON logs(timestamp);
CREATE INDEX idx_logs_module ON logs(module);
CREATE INDEX idx_logs_level ON logs(level);
CREATE INDEX idx_logs_module_level ON logs(module, level);
CREATE INDEX idx_logs_timestamp_module ON logs(timestamp, module);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_logs_timestamp_module;
DROP INDEX IF EXISTS idx_logs_module_level;
DROP INDEX IF EXISTS idx_logs_level;
DROP INDEX IF EXISTS idx_logs_module;
DROP INDEX IF EXISTS idx_logs_timestamp;
DROP TABLE logs;
-- +goose StatementEnd
