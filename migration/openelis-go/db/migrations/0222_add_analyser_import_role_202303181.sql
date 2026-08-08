-- source: liquibase liquibase/2.7.x.x/add_analyser_import_role.xml::202303181::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Attach importAnalyzer url path to Analyzer result module
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/importAnalyzer', (SELECT id FROM clinlims.system_module WHERE name = 'AnalyzerResults')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_analyser_import_role.xml::202303181::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
