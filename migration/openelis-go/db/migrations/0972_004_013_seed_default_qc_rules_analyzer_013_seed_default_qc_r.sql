-- source: liquibase liquibase/analyzer/004-013-seed-default-qc-rules.xml::analyzer-013-seed-default-qc-rules::fr-15-qc-rules
-- +goose Up
-- +goose StatementBegin
-- FR-15: no-op — QC rules come from analyzer profiles now (see configDefaults.qcRules)
SELECT 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/analyzer/004-013-seed-default-qc-rules.xml::analyzer-013-seed-default-qc-rules::fr-15-qc-rules
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
