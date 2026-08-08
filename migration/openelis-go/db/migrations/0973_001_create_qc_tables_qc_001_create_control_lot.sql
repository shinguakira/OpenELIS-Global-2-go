-- source: liquibase liquibase/qc/001-create-qc-tables.xml::qc-001-create-control-lot::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create QC control lot table for managing quality control material batches
CREATE TABLE IF NOT EXISTS qc_control_lot (id VARCHAR(36) NOT NULL, fhir_uuid UUID DEFAULT gen_random_uuid() NOT NULL, product_name VARCHAR(255) NOT NULL, lot_number VARCHAR(100) NOT NULL, manufacturer VARCHAR(255), control_level VARCHAR(50) NOT NULL, test_id INTEGER NOT NULL, instrument_id INTEGER NOT NULL, calculation_method VARCHAR(50) DEFAULT 'INITIAL_RUNS' NOT NULL, initial_runs_count INTEGER DEFAULT 20, manufacturer_mean numeric(15, 5), manufacturer_std_dev numeric(15, 5), activation_date TIMESTAMP WITHOUT TIME ZONE, expiration_date TIMESTAMP WITHOUT TIME ZONE, status VARCHAR(50) DEFAULT 'ESTABLISHMENT' NOT NULL, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qc_control_lot_pkey PRIMARY KEY (id), UNIQUE (fhir_uuid));
ALTER TABLE qc_control_lot ADD CONSTRAINT fk_qc_control_lot_test FOREIGN KEY (test_id) REFERENCES test (id);
ALTER TABLE qc_control_lot ADD CONSTRAINT fk_qc_control_lot_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_qc_control_lot_test_instrument ON qc_control_lot(test_id, instrument_id);
CREATE INDEX IF NOT EXISTS idx_qc_control_lot_status ON qc_control_lot(status);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/001-create-qc-tables.xml::qc-001-create-control-lot::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
