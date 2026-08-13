-- source: liquibase liquibase/2.3.x.x/accession_validation_range_access.xml::1::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/AccessionValidationRange', (SELECT id FROM clinlims.system_module WHERE name = 'ResultsValidationGeneral')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/accession_validation_range_access.xml::1::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
