-- source: liquibase liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-04-create-analyzer-pk::ogc-492
-- +goose Up
-- +goose StatementBegin
-- OGC-492: New PK keyed by analyzer instance, not type
ALTER TABLE analyzer_test_map ALTER COLUMN  analyzer_id SET NOT NULL;

ALTER TABLE analyzer_test_map ADD CONSTRAINT analyzer_test_map_pk PRIMARY KEY (analyzer_id, analyzer_test_name);

ALTER TABLE analyzer_test_map ADD CONSTRAINT analyzer_test_map_analyzer_fk FOREIGN KEY (analyzer_id) REFERENCES analyzer (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/006-per-analyzer-test-mappings.xml::013-006-04-create-analyzer-pk::ogc-492
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
