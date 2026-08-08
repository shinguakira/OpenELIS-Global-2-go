-- source: liquibase liquibase/qc/001-create-qc-tables.xml::qc-003-create-qc-statistics::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create QC statistics table for caching calculated statistical parameters
CREATE TABLE IF NOT EXISTS qc_statistics (id VARCHAR(36) NOT NULL, control_lot_id VARCHAR(36) NOT NULL, calculation_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, mean numeric(15, 5) NOT NULL, standard_deviation numeric(15, 5) NOT NULL, num_values INTEGER NOT NULL, calculation_method VARCHAR(50) NOT NULL, validity_start TIMESTAMP WITHOUT TIME ZONE NOT NULL, validity_end TIMESTAMP WITHOUT TIME ZONE, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qc_statistics_pkey PRIMARY KEY (id));
ALTER TABLE qc_statistics ADD CONSTRAINT fk_qc_statistics_control_lot FOREIGN KEY (control_lot_id) REFERENCES qc_control_lot (id);
ALTER TABLE qc_statistics ADD CONSTRAINT fk_qc_statistics_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_qc_statistics_lot_date ON qc_statistics(control_lot_id, calculation_date DESC);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/001-create-qc-tables.xml::qc-003-create-qc-statistics::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
