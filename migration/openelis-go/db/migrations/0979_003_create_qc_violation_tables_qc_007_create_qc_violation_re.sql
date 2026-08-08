-- source: liquibase liquibase/qc/003-create-qc-violation-tables.xml::qc-007-create-qc-violation-results::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create junction table linking violations to all involved QC results
CREATE TABLE IF NOT EXISTS qc_violation_results (id VARCHAR(36) NOT NULL, violation_id VARCHAR(36) NOT NULL, result_id VARCHAR(36) NOT NULL, CONSTRAINT qc_violation_results_pkey PRIMARY KEY (id));
ALTER TABLE qc_violation_results ADD CONSTRAINT fk_violation_results_violation FOREIGN KEY (violation_id) REFERENCES qc_rule_violation (id) ON DELETE CASCADE;
ALTER TABLE qc_violation_results ADD CONSTRAINT fk_violation_results_result FOREIGN KEY (result_id) REFERENCES qc_result (id);
CREATE INDEX IF NOT EXISTS idx_violation_results_violation ON qc_violation_results(violation_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/003-create-qc-violation-tables.xml::qc-007-create-qc-violation-results::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
