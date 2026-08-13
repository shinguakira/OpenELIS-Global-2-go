-- source: liquibase liquibase/analyzer/004-012-create-analyzer-qc-rule.xml::analyzer-012-create-analyzer-qc-rule::fr-15-qc-rules
-- +goose Up
-- +goose StatementBegin
-- FR-15: QC Sample Identification Rules table
CREATE TABLE IF NOT EXISTS analyzer_qc_rule (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, rule_type VARCHAR(30) NOT NULL, target_field VARCHAR(100), operand VARCHAR(500) NOT NULL, is_active BOOLEAN DEFAULT TRUE NOT NULL, display_order INTEGER DEFAULT 0 NOT NULL, description VARCHAR(255), sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_qc_rule_pkey PRIMARY KEY (id), CONSTRAINT fk_analyzer_qc_rule_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id) ON DELETE CASCADE);
ALTER TABLE analyzer_qc_rule
            ADD CONSTRAINT chk_qc_rule_type
            CHECK (rule_type IN ('FIELD_EQUALS', 'SPECIMEN_ID_PREFIX', 'SPECIMEN_ID_PATTERN', 'FIELD_CONTAINS'));
CREATE INDEX IF NOT EXISTS idx_analyzer_qc_rule_analyzer_id ON analyzer_qc_rule(analyzer_id);
CREATE INDEX IF NOT EXISTS idx_analyzer_qc_rule_active ON analyzer_qc_rule(analyzer_id, is_active);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/analyzer/004-012-create-analyzer-qc-rule.xml::analyzer-012-create-analyzer-qc-rule::fr-15-qc-rules
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
