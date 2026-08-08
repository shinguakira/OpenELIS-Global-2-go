-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-06-delete-phantom-analyzer-rows::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Delete phantom analyzer rows (no IP, no port) that were auto-created by plugin connect()
DELETE FROM analyzer
            WHERE ip_address IS NULL
            AND port IS NULL
            AND analyzer_type_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-06-delete-phantom-analyzer-rows::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
