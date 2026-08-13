-- source: liquibase liquibase/3.4.14.x/011-complete-file-config-unification.xml::011-migrate-fic-data-to-analyzer::pmanko
-- +goose Up
-- +goose StatementBegin
-- Migrate existing data from file_import_configuration to analyzer.
--         Skipped if the FIC table was never created.
UPDATE clinlims.analyzer a
            SET delimiter = fic.delimiter,
                has_header = fic.has_header,
                skip_rows = fic.skip_rows,
                archive_directory = fic.archive_directory,
                error_directory = fic.error_directory
            FROM clinlims.file_import_configuration fic
            WHERE fic.analyzer_id = a.id::int;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/011-complete-file-config-unification.xml::011-migrate-fic-data-to-analyzer::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
