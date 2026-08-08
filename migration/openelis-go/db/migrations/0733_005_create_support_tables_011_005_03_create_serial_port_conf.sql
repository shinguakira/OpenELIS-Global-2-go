-- source: liquibase liquibase/3.4.x.x/005-create-support-tables.xml::011-005-03-create-serial-port-configuration-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create serial_port_configuration table for RS232 communication parameters
CREATE TABLE IF NOT EXISTS serial_port_configuration (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, port_name VARCHAR(50) NOT NULL, baud_rate INTEGER DEFAULT 9600 NOT NULL, data_bits INTEGER DEFAULT 8 NOT NULL, stop_bits VARCHAR(10) DEFAULT 'ONE' NOT NULL, parity VARCHAR(10) DEFAULT 'NONE' NOT NULL, flow_control VARCHAR(20) DEFAULT 'NONE' NOT NULL, active BOOLEAN DEFAULT TRUE NOT NULL, fhir_uuid UUID NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT serial_port_configuration_pkey PRIMARY KEY (id), CONSTRAINT fk_serial_port_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id) ON DELETE CASCADE, UNIQUE (analyzer_id), UNIQUE (fhir_uuid));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS serial_port_configuration;
-- +goose StatementEnd
