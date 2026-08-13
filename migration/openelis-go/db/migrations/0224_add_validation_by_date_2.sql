-- source: liquibase liquibase/2.7.x.x/add_validation_by_date.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add ResultValidationByTestDate module url to the ResultsValidationGeneral module
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ResultValidationByTestDate', (SELECT id FROM clinlims.system_module WHERE name = 'ResultsValidationGeneral')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_validation_by_date.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
