-- source: liquibase liquibase/3.5.x.x/049-test-alert-rule.xml::OGC-763-test-alert-rule-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_alert_rule (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, name VARCHAR(100) NOT NULL, is_enabled BOOLEAN DEFAULT TRUE NOT NULL, trigger_type VARCHAR(30) NOT NULL, trigger_value VARCHAR(100), notify_sms BOOLEAN DEFAULT FALSE NOT NULL, notify_email BOOLEAN DEFAULT FALSE NOT NULL, notify_ordering_physician BOOLEAN DEFAULT FALSE NOT NULL, notify_patient BOOLEAN DEFAULT FALSE NOT NULL, notify_referring_facility BOOLEAN DEFAULT FALSE NOT NULL, notify_custom_phone VARCHAR(20), notify_custom_email VARCHAR(100), notify_role_id numeric(10), acknowledgment_required BOOLEAN DEFAULT FALSE NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_alert_rule PRIMARY KEY (id));
ALTER TABLE clinlims.test_alert_rule
            ADD CONSTRAINT test_alert_rule_trigger_chk
            CHECK (trigger_type IN ('ALL', 'ABNORMAL', 'CRITICAL', 'SPECIFIC_VALUE', 'COMPLIANCE_BREACH'));
ALTER TABLE clinlims.test_alert_rule ADD CONSTRAINT fk_test_alert_rule_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
ALTER TABLE clinlims.test_alert_rule ADD CONSTRAINT fk_test_alert_rule_role FOREIGN KEY (notify_role_id) REFERENCES clinlims.system_role (id);
CREATE INDEX IF NOT EXISTS idx_test_alert_rule_test ON clinlims.test_alert_rule(test_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/049-test-alert-rule.xml::OGC-763-test-alert-rule-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
