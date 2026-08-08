-- source: liquibase liquibase/3.4.14.x/007-unify-analyzer-transport-config.xml::013-007-02-migrate-file-config-to-analyzer::ogc-492
-- +goose Up
-- +goose StatementBegin
-- Migrate FILE config from file_import_configuration to analyzer table
UPDATE analyzer a
            SET import_directory = fic.import_directory,
                file_pattern = fic.file_pattern,
                column_mappings_json = fic.column_mappings,
                file_format = fic.file_format
            FROM file_import_configuration fic
            WHERE fic.analyzer_id = a.id;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/007-unify-analyzer-transport-config.xml::013-007-02-migrate-file-config-to-analyzer::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
