-- source: liquibase liquibase/qc/001-create-qc-tables.xml::qc-002-create-qc-result::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create QC result table for storing quality control measurements
CREATE TABLE IF NOT EXISTS qc_result (id VARCHAR(36) NOT NULL, control_lot_id VARCHAR(36) NOT NULL, test_id INTEGER NOT NULL, instrument_id INTEGER NOT NULL, result_value numeric(15, 5) NOT NULL, unit_of_measure VARCHAR(50) NOT NULL, z_score numeric(10, 4), run_date_time TIMESTAMP WITHOUT TIME ZONE NOT NULL, technician_id INTEGER, result_status VARCHAR(50) DEFAULT 'PENDING' NOT NULL, non_conformity_flag BOOLEAN DEFAULT FALSE, external_notes TEXT, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qc_result_pkey PRIMARY KEY (id));
ALTER TABLE qc_result ADD CONSTRAINT fk_qc_result_control_lot FOREIGN KEY (control_lot_id) REFERENCES qc_control_lot (id);
ALTER TABLE qc_result ADD CONSTRAINT fk_qc_result_test FOREIGN KEY (test_id) REFERENCES test (id);
ALTER TABLE qc_result ADD CONSTRAINT fk_qc_result_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_qc_result_control_lot_date ON qc_result(control_lot_id, run_date_time DESC);
CREATE INDEX IF NOT EXISTS idx_qc_result_instrument_date ON qc_result(instrument_id, run_date_time DESC);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/001-create-qc-tables.xml::qc-002-create-qc-result::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
