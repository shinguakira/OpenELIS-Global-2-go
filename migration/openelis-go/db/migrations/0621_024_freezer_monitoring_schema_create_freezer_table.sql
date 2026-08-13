-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create freezer table for storing cold storage device monitoring configuration
CREATE SEQUENCE  IF NOT EXISTS clinlims.freezer_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.freezer (id BIGINT NOT NULL, name VARCHAR(128) NOT NULL, protocol VARCHAR(8) NOT NULL, host VARCHAR(255), port INTEGER, serial_port VARCHAR(255), baud_rate INTEGER, data_bits INTEGER, stop_bits INTEGER, parity VARCHAR(8), slave_id INTEGER NOT NULL, temperature_register INTEGER NOT NULL, humidity_register INTEGER, temperature_scale DECIMAL DEFAULT 1, temperature_offset DECIMAL DEFAULT 0, humidity_scale DECIMAL DEFAULT 1, humidity_offset DECIMAL DEFAULT 0, target_temperature DECIMAL, warning_threshold DECIMAL, critical_threshold DECIMAL, polling_interval_seconds INTEGER DEFAULT 60, active BOOLEAN DEFAULT TRUE, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT freezer_pkey PRIMARY KEY (id), UNIQUE (name));
CREATE UNIQUE INDEX IF NOT EXISTS idx_freezer_name ON clinlims.freezer(name);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
