-- +goose Up
-- Agregar campo created_at a trade_links
ALTER TABLE trade_links ADD COLUMN created_at DATETIME;

-- Actualizar registros existentes con timestamp actual
UPDATE trade_links SET created_at = datetime('now') WHERE created_at IS NULL;

-- +goose Down
-- SQLite no soporta DROP COLUMN directamente, pero podemos dejar el campo
-- En caso de rollback, simplemente ignoramos el campo created_at
