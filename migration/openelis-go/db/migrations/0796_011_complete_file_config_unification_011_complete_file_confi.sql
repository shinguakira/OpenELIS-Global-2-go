-- source: liquibase liquibase/3.4.14.x/011-complete-file-config-unification.xml::011-complete-file-config-unification::pmanko
-- +goose Up
-- +goose StatementBegin
-- Complete FILE config unification on analyzer table. Adds remaining fields
--         from file_import_configuration (delimiter, has_header, skip_rows, archive_directory,
--         error_directory) so FIC can be fully deprecated.
ALTER TABLE analyzer ADD IF NOT EXISTS delimiter VARCHAR(10) DEFAULT ',';
ALTER TABLE analyzer ADD IF NOT EXISTS has_header BOOLEAN DEFAULT TRUE;
ALTER TABLE analyzer ADD IF NOT EXISTS skip_rows INTEGER DEFAULT 0;
ALTER TABLE analyzer ADD IF NOT EXISTS archive_directory VARCHAR(500);
ALTER TABLE analyzer ADD IF NOT EXISTS error_directory VARCHAR(500);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/011-complete-file-config-unification.xml::011-complete-file-config-unification::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
