-- source: liquibase liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-03-drop-type-pk::ogc-492
-- +goose Up
-- +goose StatementBegin
-- OGC-492: Drop type-based PK and FK
ALTER TABLE analyzer_test_map DROP CONSTRAINT analyzer_test_map_pk;

ALTER TABLE analyzer_test_map DROP CONSTRAINT IF EXISTS analyzer_test_map_analyzer_type_fk;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-03-drop-type-pk::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
