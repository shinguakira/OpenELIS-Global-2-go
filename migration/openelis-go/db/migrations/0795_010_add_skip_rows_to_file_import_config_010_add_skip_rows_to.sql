-- source: liquibase liquibase/3.4.14.x/010-add-skip-rows-to-file-import-config.xml::010-add-skip-rows-to-file-import-config::pmanko
-- +goose Up
-- +goose StatementBegin
-- Add skip_rows column for instruments with metadata lines before headers (e.g. Wondfo Finecare)
ALTER TABLE file_import_configuration ADD IF NOT EXISTS skip_rows INTEGER DEFAULT 0;
UPDATE file_import_configuration SET skip_rows = 0 WHERE skip_rows IS NULL;
ALTER TABLE file_import_configuration ALTER COLUMN  skip_rows SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/010-add-skip-rows-to-file-import-config.xml::010-add-skip-rows-to-file-import-config::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
