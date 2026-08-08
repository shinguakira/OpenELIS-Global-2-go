-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-alert-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create generic alert table for system-wide alert management (freezer monitoring, equipment monitoring, inventory alerts, etc.)
CREATE SEQUENCE  IF NOT EXISTS clinlims.alert_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.alert (id BIGINT NOT NULL, alert_type VARCHAR(50) NOT NULL, alert_entity_type VARCHAR(100) NOT NULL, alert_entity_id BIGINT NOT NULL, severity VARCHAR(20) NOT NULL, status VARCHAR(20) NOT NULL, start_time TIMESTAMP WITH TIME ZONE NOT NULL, end_time TIMESTAMP WITH TIME ZONE, message TEXT NOT NULL, context_data JSONB, acknowledged_at TIMESTAMP WITH TIME ZONE, acknowledged_by INTEGER, resolved_at TIMESTAMP WITH TIME ZONE, resolved_by INTEGER, resolution_notes TEXT, last_duplicate_time TIMESTAMP WITH TIME ZONE, duplicate_count INTEGER DEFAULT 0 NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT alert_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.alert ADD CONSTRAINT fk_alert_acknowledged_by FOREIGN KEY (acknowledged_by) REFERENCES clinlims.system_user (id);
ALTER TABLE clinlims.alert ADD CONSTRAINT fk_alert_resolved_by FOREIGN KEY (resolved_by) REFERENCES clinlims.system_user (id);
ALTER TABLE clinlims.alert ADD CONSTRAINT chk_alert_severity
            CHECK (severity IN ('WARNING', 'CRITICAL'));
ALTER TABLE clinlims.alert ADD CONSTRAINT chk_alert_status
            CHECK (status IN ('OPEN', 'ACKNOWLEDGED', 'RESOLVED'));
ALTER TABLE clinlims.alert ADD CONSTRAINT chk_alert_type
            CHECK (alert_type IN ('FREEZER_TEMPERATURE', 'EQUIPMENT_FAILURE', 'INVENTORY_LOW', 'SAMPLE_TRACKING', 'OTHER'));
CREATE INDEX IF NOT EXISTS idx_alert_entity ON clinlims.alert(alert_entity_type, alert_entity_id);
CREATE INDEX IF NOT EXISTS idx_alert_type ON clinlims.alert(alert_type);
CREATE INDEX IF NOT EXISTS idx_alert_status ON clinlims.alert(status);
CREATE INDEX IF NOT EXISTS idx_alert_severity ON clinlims.alert(severity);
CREATE INDEX IF NOT EXISTS idx_alert_start_time ON clinlims.alert(start_time);
CREATE INDEX IF NOT EXISTS idx_alert_status_severity ON clinlims.alert(status, severity);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-alert-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
