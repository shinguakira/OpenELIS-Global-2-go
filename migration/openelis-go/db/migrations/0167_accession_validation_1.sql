-- source: liquibase liquibase/2.3.x.x/accession_validation.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_module (id, name, description) VALUES (nextval('clinlims.system_module_seq'), 'ResultsValidationGeneral', 'General result validation pages') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_role_module (id, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), (SELECT id FROM clinlims.system_role WHERE name = 'Validator'), (SELECT id FROM clinlims.system_module WHERE name = 'ResultsValidationGeneral')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/AccessionValidation', (SELECT id FROM clinlims.system_module WHERE name = 'ResultsValidationGeneral')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/accession_validation.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
