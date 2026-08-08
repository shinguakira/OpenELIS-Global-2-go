-- source: liquibase liquibase/3.4.14.x/008-analyzer-test-map-cascade-delete.xml::008-1-analyzer-test-map-cascade-delete::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer_test_map DROP CONSTRAINT analyzer_test_map_analyzer_fk;

ALTER TABLE clinlims.analyzer_test_map ADD CONSTRAINT analyzer_test_map_analyzer_fk FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer (id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/008-analyzer-test-map-cascade-delete.xml::008-1-analyzer-test-map-cascade-delete::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
