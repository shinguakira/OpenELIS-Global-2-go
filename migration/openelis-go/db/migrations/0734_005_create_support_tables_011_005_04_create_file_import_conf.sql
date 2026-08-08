-- source: liquibase liquibase/3.4.x.x/005-create-support-tables.xml::011-005-04-create-file-import-configuration-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create file_import_configuration table for CSV/file-based result import
CREATE TABLE IF NOT EXISTS file_import_configuration (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, import_directory VARCHAR(255) NOT NULL, file_pattern VARCHAR(100) DEFAULT '*.csv' NOT NULL, archive_directory VARCHAR(255), error_directory VARCHAR(255), column_mappings TEXT, delimiter VARCHAR(10) DEFAULT ',' NOT NULL, has_header BOOLEAN DEFAULT TRUE NOT NULL, active BOOLEAN DEFAULT TRUE NOT NULL, fhir_uuid UUID NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT file_import_configuration_pkey PRIMARY KEY (id), CONSTRAINT fk_file_import_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id) ON DELETE CASCADE, UNIQUE (analyzer_id), UNIQUE (fhir_uuid));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS file_import_configuration;
-- +goose StatementEnd
