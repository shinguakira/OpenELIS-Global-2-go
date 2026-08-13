-- source: liquibase liquibase/qc/002-create-westgard-rule-config.xml::qc-004-create-westgard-rule-config::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create Westgard rule configuration table for managing QC rule enablement per test-instrument
CREATE TABLE IF NOT EXISTS westgard_rule_config (id VARCHAR(36) NOT NULL, test_id INTEGER NOT NULL, instrument_id INTEGER NOT NULL, rule_code VARCHAR(20) NOT NULL, enabled BOOLEAN DEFAULT TRUE NOT NULL, severity VARCHAR(20) NOT NULL, requires_corrective_action BOOLEAN DEFAULT FALSE NOT NULL, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT westgard_rule_config_pkey PRIMARY KEY (id));
ALTER TABLE westgard_rule_config ADD CONSTRAINT fk_westgard_rule_test FOREIGN KEY (test_id) REFERENCES test (id);
ALTER TABLE westgard_rule_config ADD CONSTRAINT fk_westgard_rule_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
ALTER TABLE westgard_rule_config ADD CONSTRAINT uk_westgard_rule_test_instrument_rule UNIQUE (test_id, instrument_id, rule_code);
CREATE INDEX IF NOT EXISTS idx_westgard_rule_instrument ON westgard_rule_config(instrument_id);
CREATE INDEX IF NOT EXISTS idx_westgard_rule_enabled ON westgard_rule_config(enabled);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/002-create-westgard-rule-config.xml::qc-004-create-westgard-rule-config::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
