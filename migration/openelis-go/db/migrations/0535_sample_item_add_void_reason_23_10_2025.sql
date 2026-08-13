-- source: liquibase liquibase/3.2.x.x/sample_item.xml::add-void-reason-23-10-2025::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add void_reason column to sample_item table
ALTER TABLE sample_item ADD IF NOT EXISTS void_reason VARCHAR(200);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sample_item DROP COLUMN IF EXISTS void_reason;
-- +goose StatementEnd
