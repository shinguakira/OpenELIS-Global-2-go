-- source: liquibase liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-002-add-report-type::Agaba_derrick
-- +goose Up
-- +goose StatementBegin
-- Add report_type to report_definition for filtering
ALTER TABLE clinlims.report_definition ADD IF NOT EXISTS report_type VARCHAR(50);
CREATE INDEX IF NOT EXISTS idx_report_def_report_type ON clinlims.report_definition(report_type);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-002-add-report-type::Agaba_derrick
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
