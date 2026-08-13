-- source: liquibase liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-foreign-keys::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test_method ADD CONSTRAINT test_method_test_id_fkey FOREIGN KEY (test_id) REFERENCES clinlims.test (id) ON DELETE CASCADE;

ALTER TABLE clinlims.test_method ADD CONSTRAINT test_method_method_id_fkey FOREIGN KEY (method_id) REFERENCES clinlims.method (id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-foreign-keys::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
