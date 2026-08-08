-- source: liquibase liquibase/2.7.x.x/add_loinc_code.xml::23102501::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test SET loinc = '25836-8' WHERE id=(SELECT id FROM test WHERE name = 'Viral Load');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_loinc_code.xml::23102501::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
