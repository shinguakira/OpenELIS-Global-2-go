-- source: liquibase liquibase/3.5.x.x/051-uom-code-ucum.xml::uom-add-code-ucum::openelis
-- +goose Up
-- +goose StatementBegin
-- Add optional code + UCUM code columns to unit_of_measure for the inline unit-create form.
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS code VARCHAR(50);
ALTER TABLE clinlims.unit_of_measure ADD IF NOT EXISTS ucum_code VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/051-uom-code-ucum.xml::uom-add-code-ucum::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
