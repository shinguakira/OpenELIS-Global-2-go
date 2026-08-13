-- source: liquibase liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-06-add-max-fields-config::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add analyzer max fields per query configuration (default 500)
INSERT INTO clinlims.system_configuration (id, name, value, description, active) VALUES (nextval('clinlims.system_configuration_seq'), 'analyzer.max.fields.per.query', '500', 'Maximum fields returned per query', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-06-add-max-fields-config::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
