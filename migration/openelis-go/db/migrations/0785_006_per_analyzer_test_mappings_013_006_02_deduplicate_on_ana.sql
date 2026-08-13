-- source: liquibase liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-02-deduplicate-on-analyzer-id::ogc-492
-- +goose Up
-- +goose StatementBegin
-- OGC-492: Deduplicate mappings before creating per-analyzer PK
DELETE FROM analyzer_test_map WHERE analyzer_id IS NULL;

DELETE FROM analyzer_test_map a
            USING analyzer_test_map b
            WHERE a.ctid < b.ctid
            AND a.analyzer_id = b.analyzer_id
            AND a.analyzer_test_name = b.analyzer_test_name;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-02-deduplicate-on-analyzer-id::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
