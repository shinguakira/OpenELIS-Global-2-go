-- source: liquibase liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-04-add-query-timeout-config::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add analyzer query timeout configuration (default 5 minutes)
INSERT INTO clinlims.system_configuration (id, name, value, description, active) VALUES (nextval('clinlims.system_configuration_seq'), 'analyzer.query.timeout.minutes', '5', 'Query analyzer timeout in minutes', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-04-add-query-timeout-config::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
