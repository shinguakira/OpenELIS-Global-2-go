-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::remove-system-config-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Remove system_config table - migrated to SiteInformation (B1 remediation)
DROP TABLE IF EXISTS clinlims.system_config;
DROP SEQUENCE IF EXISTS clinlims.system_config_seq CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::remove-system-config-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
