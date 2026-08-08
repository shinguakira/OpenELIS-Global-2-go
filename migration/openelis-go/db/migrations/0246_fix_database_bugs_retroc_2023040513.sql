-- source: liquibase liquibase/2.7.x.x/fix_database_bugs_retroc.xml::2023040513::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, has_delete, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'Y', 'Y', 'Y', (SELECT id FROM system_role WHERE name = 'Reception'), (SELECT id FROM system_module WHERE name = 'PatientEditByProject:readonly' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/fix_database_bugs_retroc.xml::2023040513::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
