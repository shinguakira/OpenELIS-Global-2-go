-- source: liquibase liquibase/3.4.x.x/003-create-field-tables.xml::011-003-01-create-analyzer-field-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_field table with custom_field_type FK baked in
CREATE TABLE IF NOT EXISTS analyzer_field (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, field_name VARCHAR(255) NOT NULL, astm_ref VARCHAR(50), field_type VARCHAR(20) NOT NULL, unit VARCHAR(50), custom_field_type_id VARCHAR(36), is_active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_field_pkey PRIMARY KEY (id), CONSTRAINT fk_analyzer_field_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id), CONSTRAINT fk_analyzer_field_custom_type FOREIGN KEY (custom_field_type_id) REFERENCES custom_field_type(id));
ALTER TABLE clinlims.analyzer_field
            ADD CONSTRAINT chk_field_type
            CHECK (field_type IN ('NUMERIC', 'QUALITATIVE', 'CONTROL_TEST', 'MELTING_POINT', 'DATE_TIME', 'TEXT', 'CUSTOM'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/003-create-field-tables.xml::011-003-01-create-analyzer-field-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
