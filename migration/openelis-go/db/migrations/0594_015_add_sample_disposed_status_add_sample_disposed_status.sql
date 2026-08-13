-- source: liquibase liquibase/3.3.x.x/015-add-sample-disposed-status.xml::add-sample-disposed-status::ogc-73
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.status_of_sample (id, description, code, status_type, lastupdated, name, display_key, is_active) VALUES (nextval('clinlims.status_of_sample_seq'), 'Sample has been physically disposed', '1', 'SAMPLE', NOW(), 'SampleDisposed', 'status.sample.disposed', 'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/015-add-sample-disposed-status.xml::add-sample-disposed-status::ogc-73
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
