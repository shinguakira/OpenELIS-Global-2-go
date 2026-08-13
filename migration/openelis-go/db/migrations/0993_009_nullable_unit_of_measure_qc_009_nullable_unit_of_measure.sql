-- source: liquibase liquibase/qc/009-nullable-unit-of-measure.xml::qc-009-nullable-unit-of-measure::mvp-flow-fix
-- +goose Up
-- +goose StatementBegin
-- Make unit_of_measure nullable — qualitative PCR results have no units
ALTER TABLE clinlims.qc_result ALTER COLUMN  unit_of_measure DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/009-nullable-unit-of-measure.xml::qc-009-nullable-unit-of-measure::mvp-flow-fix
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
