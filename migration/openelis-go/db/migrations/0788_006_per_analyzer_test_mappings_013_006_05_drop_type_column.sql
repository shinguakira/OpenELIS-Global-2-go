-- source: liquibase liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-05-drop-type-column::ogc-492
-- +goose Up
-- +goose StatementBegin
-- OGC-492: analyzer_type has no role in test mappings — drop from table
ALTER TABLE analyzer_test_map DROP COLUMN IF EXISTS analyzer_type_id;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-05-drop-type-column::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
