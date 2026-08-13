-- source: liquibase liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-002-add-is-public::Agaba_derrick
-- +goose Up
-- +goose StatementBegin
-- Add is_public to report_definition for visibility
ALTER TABLE clinlims.report_definition ADD IF NOT EXISTS is_public BOOLEAN DEFAULT TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.report_definition DROP COLUMN IF EXISTS is_public;
-- +goose StatementEnd
