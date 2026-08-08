-- source: liquibase liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-001-create-report-definition-table::Agaba-derrick
-- +goose Up
-- +goose StatementBegin
-- Create report_definition table for storing user-created report definitions
CREATE TABLE IF NOT EXISTS clinlims.report_definition (id VARCHAR(50) NOT NULL, name VARCHAR(200) NOT NULL, description TEXT, category VARCHAR(100), definition_json TEXT NOT NULL, created_by VARCHAR(50), created_date TIMESTAMP WITHOUT TIME ZONE, last_updated TIMESTAMP WITHOUT TIME ZONE, is_active BOOLEAN DEFAULT TRUE NOT NULL, CONSTRAINT report_definition_pkey PRIMARY KEY (id));
CREATE INDEX IF NOT EXISTS idx_report_def_category ON clinlims.report_definition(category);
CREATE INDEX IF NOT EXISTS idx_report_def_created_by ON clinlims.report_definition(created_by);
CREATE INDEX IF NOT EXISTS idx_report_def_is_active ON clinlims.report_definition(is_active);
CREATE INDEX IF NOT EXISTS idx_report_def_last_updated ON clinlims.report_definition(last_updated);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-001-create-report-definition-table::Agaba-derrick
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
