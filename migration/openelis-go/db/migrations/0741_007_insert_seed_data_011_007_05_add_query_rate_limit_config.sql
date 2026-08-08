-- source: liquibase liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-05-add-query-rate-limit-config::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add analyzer query rate limit configuration (default 1 per minute)
INSERT INTO clinlims.system_configuration (id, name, value, description, active) VALUES (nextval('clinlims.system_configuration_seq'), 'analyzer.query.rate.limit.per.minute', '1', 'Maximum queries per analyzer per minute', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-05-add-query-rate-limit-config::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
