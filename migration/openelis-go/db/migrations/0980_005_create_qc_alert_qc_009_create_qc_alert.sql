-- source: liquibase liquibase/qc/005-create-qc-alert.xml::qc-009-create-qc-alert::openelisglobal
-- +goose Up
-- +goose StatementBegin
-- Create QC alert table for tracking notifications sent about violations
CREATE TABLE IF NOT EXISTS qc_alert (id VARCHAR(36) NOT NULL, violation_id VARCHAR(36) NOT NULL, alert_type VARCHAR(50) NOT NULL, recipient_user_id INTEGER NOT NULL, recipient_email VARCHAR(255), sent_date_time TIMESTAMP WITHOUT TIME ZONE NOT NULL, read_status BOOLEAN DEFAULT FALSE NOT NULL, read_date_time TIMESTAMP WITHOUT TIME ZONE, message_subject VARCHAR(500), message_body TEXT, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT qc_alert_pkey PRIMARY KEY (id));
ALTER TABLE qc_alert ADD CONSTRAINT fk_qc_alert_violation FOREIGN KEY (violation_id) REFERENCES qc_rule_violation (id);
ALTER TABLE qc_alert ADD CONSTRAINT fk_qc_alert_recipient FOREIGN KEY (recipient_user_id) REFERENCES system_user (id);
ALTER TABLE qc_alert ADD CONSTRAINT fk_qc_alert_sys_user FOREIGN KEY (sys_user_id) REFERENCES system_user (id);
CREATE INDEX IF NOT EXISTS idx_qc_alert_violation ON qc_alert(violation_id);
CREATE INDEX IF NOT EXISTS idx_qc_alert_recipient ON qc_alert(recipient_user_id);
CREATE INDEX IF NOT EXISTS idx_qc_alert_read_status ON qc_alert(read_status);
CREATE INDEX IF NOT EXISTS idx_qc_alert_sent_date ON qc_alert(sent_date_time DESC);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/qc/005-create-qc-alert.xml::qc-009-create-qc-alert::openelisglobal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
