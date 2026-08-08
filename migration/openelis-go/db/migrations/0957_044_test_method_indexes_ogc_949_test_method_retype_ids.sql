-- source: liquibase liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-retype-ids::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test_method ALTER COLUMN test_id TYPE numeric(10,0) USING test_id::numeric(10,0);

ALTER TABLE clinlims.test_method ALTER COLUMN method_id TYPE numeric(10,0) USING method_id::numeric(10,0);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-test-method-retype-ids::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
