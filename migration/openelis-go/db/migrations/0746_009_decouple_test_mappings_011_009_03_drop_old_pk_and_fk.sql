-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-03-drop-old-pk-and-fk::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Drop old PK (analyzer_id, analyzer_test_name) and FK to analyzer
ALTER TABLE analyzer_test_map DROP CONSTRAINT analyzer_test_map_pk;

ALTER TABLE analyzer_test_map DROP CONSTRAINT analyzer_test_map_analyzer_fk;

ALTER TABLE analyzer_test_map ALTER COLUMN  analyzer_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-03-drop-old-pk-and-fk::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
