-- source: liquibase liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-01-resolve-analyzer-id::ogc-492
-- +goose Up
-- +goose StatementBegin
-- OGC-492: Ensure every test mapping has a valid analyzer_id
INSERT INTO analyzer (id, name, analyzer_type_id, is_active, last_updated)
            SELECT nextval('analyzer_seq'),
                   at.name,
                   at.id,
                   true,
                   NOW()
            FROM analyzer_type at
            WHERE at.id IN (
                SELECT DISTINCT atm.analyzer_type_id
                FROM analyzer_test_map atm
                WHERE atm.analyzer_id IS NULL
            )
            AND NOT EXISTS (
                SELECT 1 FROM analyzer a WHERE a.analyzer_type_id = at.id
            ) ON CONFLICT DO NOTHING;
UPDATE analyzer_test_map atm
            SET analyzer_id = (
                SELECT MIN(a.id) FROM analyzer a
                WHERE a.analyzer_type_id = atm.analyzer_type_id
            )
            WHERE atm.analyzer_id IS NULL
            AND atm.analyzer_type_id IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-01-resolve-analyzer-id::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
