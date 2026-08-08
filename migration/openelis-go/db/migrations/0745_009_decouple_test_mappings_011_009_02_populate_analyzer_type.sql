-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-02-populate-analyzer-type-id::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Migrate existing test mappings: copy analyzer_type_id from parent analyzer row
UPDATE analyzer_test_map atm
            SET analyzer_type_id = a.analyzer_type_id
            FROM analyzer a
            WHERE atm.analyzer_id = a.id
            AND a.analyzer_type_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-02-populate-analyzer-type-id::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
