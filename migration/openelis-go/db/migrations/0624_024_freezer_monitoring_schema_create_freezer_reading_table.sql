-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-reading-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create freezer_reading table for storing temperature/humidity sensor readings
CREATE SEQUENCE  IF NOT EXISTS clinlims.freezer_reading_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.freezer_reading (id BIGINT NOT NULL, freezer_id BIGINT NOT NULL, recorded_at TIMESTAMP WITH TIME ZONE NOT NULL, temperature_celsius DECIMAL, humidity_percentage DECIMAL, status VARCHAR(16) NOT NULL, transmission_ok BOOLEAN DEFAULT TRUE NOT NULL, error_message TEXT, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT freezer_reading_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.freezer_reading ADD CONSTRAINT fk_freezer_reading_freezer FOREIGN KEY (freezer_id) REFERENCES clinlims.freezer (id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_freezer_reading_name_time ON clinlims.freezer_reading(freezer_id, recorded_at);
CREATE INDEX IF NOT EXISTS idx_freezer_reading_status ON clinlims.freezer_reading(status);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-reading-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
