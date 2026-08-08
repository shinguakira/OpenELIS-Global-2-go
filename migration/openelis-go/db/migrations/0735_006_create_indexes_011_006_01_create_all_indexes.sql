-- source: liquibase liquibase/3.4.x.x/006-create-indexes.xml::011-006-01-create-all-indexes::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create performance indexes for all analyzer tables
CREATE INDEX IF NOT EXISTS idx_analyzer_field_analyzer_id ON analyzer_field(analyzer_id);
CREATE INDEX IF NOT EXISTS idx_analyzer_field_is_active ON analyzer_field(is_active);
CREATE INDEX IF NOT EXISTS idx_field_mapping_analyzer_field_id ON analyzer_field_mapping(analyzer_field_id);
CREATE INDEX IF NOT EXISTS idx_field_mapping_is_active ON analyzer_field_mapping(is_active);
CREATE INDEX IF NOT EXISTS idx_field_mapping_openelis_field ON analyzer_field_mapping(openelis_field_id, openelis_field_type);
CREATE INDEX IF NOT EXISTS idx_analyzer_field_mapping_analyzer_id ON analyzer_field_mapping(analyzer_id);
CREATE INDEX IF NOT EXISTS idx_qualitative_mapping_analyzer_field_id ON qualitative_result_mapping(analyzer_field_id);
CREATE INDEX IF NOT EXISTS idx_unit_mapping_analyzer_field_id ON unit_mapping(analyzer_field_id);
CREATE INDEX IF NOT EXISTS idx_analyzer_error_analyzer_id ON analyzer_error(analyzer_id);
CREATE INDEX IF NOT EXISTS idx_analyzer_error_status ON analyzer_error(status);
CREATE INDEX IF NOT EXISTS idx_analyzer_error_error_type ON analyzer_error(error_type);
CREATE INDEX IF NOT EXISTS idx_analyzer_error_last_updated ON analyzer_error(last_updated);
CREATE INDEX IF NOT EXISTS idx_analyzer_error_severity ON analyzer_error(severity);
CREATE INDEX IF NOT EXISTS idx_validation_rule_custom_field_type_id ON validation_rule_configuration(custom_field_type_id);
CREATE INDEX IF NOT EXISTS idx_serial_port_config_analyzer ON serial_port_configuration(analyzer_id);
CREATE INDEX IF NOT EXISTS idx_serial_port_config_active ON serial_port_configuration(active);
CREATE INDEX IF NOT EXISTS idx_file_import_config_analyzer ON file_import_configuration(analyzer_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/006-create-indexes.xml::011-006-01-create-all-indexes::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
