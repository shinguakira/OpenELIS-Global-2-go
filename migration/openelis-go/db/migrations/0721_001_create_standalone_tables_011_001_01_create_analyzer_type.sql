-- source: liquibase liquibase/3.4.x.x/001-create-standalone-tables.xml::011-001-01-create-analyzer-type-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_type table for plugin capability definitions
CREATE TABLE IF NOT EXISTS analyzer_type (id numeric(10, 0) NOT NULL, name VARCHAR(100) NOT NULL, description VARCHAR(255), protocol VARCHAR(30) DEFAULT 'ASTM' NOT NULL, plugin_class_name VARCHAR(255), identifier_pattern VARCHAR(500), is_generic_plugin BOOLEAN DEFAULT FALSE NOT NULL, is_active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_type_pkey PRIMARY KEY (id), UNIQUE (name));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analyzer_type;
-- +goose StatementEnd
