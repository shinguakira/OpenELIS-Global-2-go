-- source: liquibase liquibase/2.3.x.x/new_tests.xml::665::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test
                SET loinc = '7855-0'
                WHERE description like 'DENGUEPCR(%)';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::665::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
