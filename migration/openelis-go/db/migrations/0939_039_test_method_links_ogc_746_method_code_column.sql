-- source: liquibase liquibase/3.5.x.x/039-test-method-links.xml::OGC-746-method-code-column::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.method ADD IF NOT EXISTS code VARCHAR(20);
ALTER TABLE clinlims.method ADD CONSTRAINT uq_method_code UNIQUE (code);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/039-test-method-links.xml::OGC-746-method-code-column::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
