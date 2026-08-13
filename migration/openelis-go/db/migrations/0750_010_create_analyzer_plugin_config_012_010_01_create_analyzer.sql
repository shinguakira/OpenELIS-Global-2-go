-- source: liquibase liquibase/3.4.x.x/010-create-analyzer-plugin-config.xml::012-010-01-create-analyzer-plugin-config::generic-astm-plugin-profiles
-- +goose Up
-- +goose StatementBegin
-- Create analyzer_plugin_config table for protocol-agnostic per-instance config defaults/overrides
CREATE TABLE IF NOT EXISTS analyzer_plugin_config (analyzer_id numeric(10, 0) NOT NULL, config JSONB DEFAULT '{}'::jsonb NOT NULL, sys_user_id VARCHAR(36), last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT analyzer_plugin_config_pkey PRIMARY KEY (analyzer_id));
ALTER TABLE analyzer_plugin_config ADD CONSTRAINT analyzer_plugin_config_analyzer_fk FOREIGN KEY (analyzer_id) REFERENCES analyzer (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/010-create-analyzer-plugin-config.xml::012-010-01-create-analyzer-plugin-config::generic-astm-plugin-profiles
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
