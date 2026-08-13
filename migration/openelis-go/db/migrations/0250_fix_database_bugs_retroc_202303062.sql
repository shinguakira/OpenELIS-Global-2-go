-- source: liquibase liquibase/2.7.x.x/fix_database_bugs_retroc.xml::202303062::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Fix roles issues while migrating from OE9.1 to OE Global
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id, system_module_param_id) VALUES (nextval('clinlims.system_module_url_seq'), '/ResultValidationByTestDate', (select id from system_module where name = 'ResultValidation' limit 1), NULL) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/fix_database_bugs_retroc.xml::202303062::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
