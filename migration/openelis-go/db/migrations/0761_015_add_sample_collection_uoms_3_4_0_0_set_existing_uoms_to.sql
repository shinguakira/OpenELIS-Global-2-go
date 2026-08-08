-- source: liquibase liquibase/3.4.x.x/015-add-sample-collection-uoms.xml::3.4.0.0-set-existing-uoms-to-result::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Set existing UOMs to RESULT type
UPDATE clinlims.unit_of_measure SET uom_type = 'RESULT' WHERE uom_type IS NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/015-add-sample-collection-uoms.xml::3.4.0.0-set-existing-uoms-to-result::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
