-- source: liquibase liquibase/qc/003-create-qc-violation-tables.xml::qc-006-create-qc-rule-violation::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create QC rule violation table for tracking Westgard rule violations
CREATE TABLE IF NOT EXISTS qc_rule_violation (id VARCHAR(36) NOT NULL, triggering_result_id VARCHAR(36) NOT NULL, rule_code VARCHAR(20) NOT NULL, violation_date_time TIMESTAMP WITHOUT TIME ZONE NOT NULL, severity VARCHAR(20) NOT NULL, instrument_id INTEGER NOT NULL, test_id INTEGER NOT NULL, resolution_status VARCHAR(50) DEFAULT 'UNRESOLVED' NOT NULL, resolved_date_time TIMESTAMP WITHOUT TIME ZONE, resolved_by_user_id INTEGER, resolution_notes TEXT, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qc_rule_violation_pkey PRIMARY KEY (id));
ALTER TABLE qc_rule_violation ADD CONSTRAINT fk_qc_violation_result FOREIGN KEY (triggering_result_id) REFERENCES qc_result (id);
ALTER TABLE qc_rule_violation ADD CONSTRAINT fk_qc_violation_test FOREIGN KEY (test_id) REFERENCES test (id);
ALTER TABLE qc_rule_violation ADD CONSTRAINT fk_qc_violation_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_qc_violation_instrument ON qc_rule_violation(instrument_id);
CREATE INDEX IF NOT EXISTS idx_qc_violation_status ON qc_rule_violation(resolution_status);
CREATE INDEX IF NOT EXISTS idx_qc_violation_severity ON qc_rule_violation(severity);
CREATE INDEX IF NOT EXISTS idx_qc_violation_date ON qc_rule_violation(violation_date_time DESC);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/003-create-qc-violation-tables.xml::qc-006-create-qc-rule-violation::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
