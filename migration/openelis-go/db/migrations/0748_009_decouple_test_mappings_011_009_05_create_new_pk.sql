-- source: liquibase liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-05-create-new-pk::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- New PK: test mappings are keyed by plugin type, not physical device
ALTER TABLE analyzer_test_map ALTER COLUMN  analyzer_type_id SET NOT NULL;

ALTER TABLE analyzer_test_map ADD CONSTRAINT analyzer_test_map_pk PRIMARY KEY (analyzer_type_id, analyzer_test_name);

ALTER TABLE analyzer_test_map ADD CONSTRAINT analyzer_test_map_analyzer_type_fk FOREIGN KEY (analyzer_type_id) REFERENCES analyzer_type (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/009-decouple-test-mappings.xml::011-009-05-create-new-pk::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
