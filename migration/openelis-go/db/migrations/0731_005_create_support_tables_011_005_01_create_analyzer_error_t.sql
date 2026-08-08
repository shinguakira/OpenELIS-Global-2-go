-- source: liquibase liquibase/3.4.x.x/005-create-support-tables.xml::011-005-01-create-analyzer-error-table::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_error table with final 7-value error_type constraint
CREATE TABLE IF NOT EXISTS analyzer_error (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, error_type VARCHAR(30) NOT NULL, severity VARCHAR(20) NOT NULL, error_message TEXT NOT NULL, raw_message TEXT, status VARCHAR(20) DEFAULT 'UNACKNOWLEDGED' NOT NULL, acknowledged_by VARCHAR(36), acknowledged_at TIMESTAMP WITHOUT TIME ZONE, resolved_at TIMESTAMP WITHOUT TIME ZONE, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_error_pkey PRIMARY KEY (id), CONSTRAINT fk_analyzer_error_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer(id));
ALTER TABLE clinlims.analyzer_error
            ADD CONSTRAINT chk_error_type
            CHECK (error_type IN ('MAPPING', 'VALIDATION', 'TIMEOUT', 'PROTOCOL', 'CONNECTION', 'QC_MAPPING_INCOMPLETE', 'QC_SERVICE_UNAVAILABLE'));
ALTER TABLE clinlims.analyzer_error
            ADD CONSTRAINT chk_severity
            CHECK (severity IN ('CRITICAL', 'ERROR', 'WARNING'));
ALTER TABLE clinlims.analyzer_error
            ADD CONSTRAINT chk_error_status
            CHECK (status IN ('UNACKNOWLEDGED', 'ACKNOWLEDGED', 'RESOLVED'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/005-create-support-tables.xml::011-005-01-create-analyzer-error-table::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
