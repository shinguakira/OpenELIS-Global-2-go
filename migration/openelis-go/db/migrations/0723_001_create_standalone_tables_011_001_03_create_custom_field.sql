-- source: liquibase liquibase/3.4.x.x/001-create-standalone-tables.xml::011-001-03-create-custom-field-type-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create custom_field_type table for extensible field type definitions
CREATE TABLE IF NOT EXISTS custom_field_type (id VARCHAR(36) NOT NULL, type_name VARCHAR(50) NOT NULL, display_name VARCHAR(100) NOT NULL, validation_pattern VARCHAR(255), value_range_min DECIMAL(10, 2), value_range_max DECIMAL(10, 2), allowed_characters VARCHAR(255), is_active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT custom_field_type_pkey PRIMARY KEY (id), UNIQUE (type_name));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS custom_field_type;
-- +goose StatementEnd
