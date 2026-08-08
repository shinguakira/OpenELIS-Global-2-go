-- source: liquibase liquibase/3.4.14.x/007-unify-analyzer-transport-config.xml::013-007-01-add-file-columns-to-analyzer::ogc-492
-- +goose Up
-- +goose StatementBegin
-- Add FILE transport columns to analyzer table for unified config
ALTER TABLE analyzer ADD IF NOT EXISTS import_directory VARCHAR(500);
ALTER TABLE analyzer ADD IF NOT EXISTS file_pattern VARCHAR(100);
ALTER TABLE analyzer ADD IF NOT EXISTS column_mappings_json TEXT;
ALTER TABLE analyzer ADD IF NOT EXISTS file_format VARCHAR(30);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/007-unify-analyzer-transport-config.xml::013-007-01-add-file-columns-to-analyzer::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
