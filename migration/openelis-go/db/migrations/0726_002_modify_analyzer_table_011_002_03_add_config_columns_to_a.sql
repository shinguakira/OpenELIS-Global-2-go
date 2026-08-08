-- source: liquibase liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-03-add-config-columns-to-analyzer::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add operational configuration columns directly to analyzer (no intermediate table)
ALTER TABLE analyzer ADD IF NOT EXISTS ip_address VARCHAR(15);
ALTER TABLE analyzer ADD IF NOT EXISTS port INTEGER;
ALTER TABLE analyzer ADD IF NOT EXISTS protocol_version VARCHAR(20) DEFAULT 'ASTM LIS2-A2';
ALTER TABLE analyzer ADD IF NOT EXISTS test_unit_ids TEXT;
ALTER TABLE analyzer ADD IF NOT EXISTS status VARCHAR(20) DEFAULT 'SETUP';
ALTER TABLE analyzer ADD IF NOT EXISTS identifier_pattern VARCHAR(255);
ALTER TABLE analyzer ADD IF NOT EXISTS last_activated_date TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-03-add-config-columns-to-analyzer::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
