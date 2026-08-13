-- source: liquibase liquibase/2.3.x.x/new_tests.xml::685::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test
                SET loinc = ''
                WHERE description like 'HIVINFANTVIRALLOAD(%)';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::685::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
