-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-04-deduplicate-mappings::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Remove orphan rows (NULL analyzer_type_id) and duplicates before creating new PK
DELETE FROM analyzer_test_map WHERE analyzer_type_id IS NULL;

DELETE FROM analyzer_test_map a
            USING analyzer_test_map b
            WHERE a.ctid < b.ctid
            AND a.analyzer_type_id = b.analyzer_type_id
            AND a.analyzer_test_name = b.analyzer_test_name;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-04-deduplicate-mappings::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
