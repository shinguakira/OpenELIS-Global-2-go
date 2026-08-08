-- source: liquibase liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-drop-unused-test-method-seq::OGC
-- +goose Up
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.test_method_seq CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/044-test-method-indexes.xml::OGC-949-drop-unused-test-method-seq::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
