-- source: liquibase liquibase/2.8.x.x/reflex_rule.xml::5::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add InternalNote and External Column to test_reflex table
ALTER TABLE clinlims.test_reflex ADD IF NOT EXISTS internal_note VARCHAR(50);
ALTER TABLE clinlims.test_reflex ADD IF NOT EXISTS external_note VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/reflex_rule.xml::5::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
