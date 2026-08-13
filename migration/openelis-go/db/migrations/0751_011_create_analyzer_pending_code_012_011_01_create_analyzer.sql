-- source: liquibase liquibase/3.4.x.x/011-create-analyzer-pending-code.xml::012-011-01-create-analyzer-pending-code::generic-astm-plugin-profiles
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_pending_code queue for auto-detected unmapped analyzer test codes
CREATE TABLE IF NOT EXISTS analyzer_pending_code (id VARCHAR(36) NOT NULL, analyzer_id numeric(10, 0) NOT NULL, analyzer_test_name VARCHAR(120) NOT NULL, first_seen_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_seen_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, seen_count INTEGER DEFAULT 1 NOT NULL, sample_payload TEXT, status VARCHAR(20) DEFAULT 'PENDING' NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_pending_code_pkey PRIMARY KEY (id));
ALTER TABLE analyzer_pending_code ADD CONSTRAINT analyzer_pending_code_analyzer_fk FOREIGN KEY (analyzer_id) REFERENCES analyzer (id);
CREATE INDEX IF NOT EXISTS idx_analyzer_pending_code_analyzer_status ON analyzer_pending_code(analyzer_id, status, last_seen_at);
CREATE INDEX IF NOT EXISTS idx_analyzer_pending_code_unique_name ON analyzer_pending_code(analyzer_id, analyzer_test_name);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/011-create-analyzer-pending-code.xml::012-011-01-create-analyzer-pending-code::generic-astm-plugin-profiles
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
