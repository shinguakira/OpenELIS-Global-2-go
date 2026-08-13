-- source: liquibase liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-06::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Add system module for EQA/QC feature
INSERT INTO clinlims.system_module (id, name, description, has_select_flag, has_add_flag, has_update_flag, has_delete_flag) VALUES (nextval('clinlims.system_module_seq'), 'EQAView', 'EQA alerts, programs, distributions, QC management', 'Y', 'Y', 'Y', 'N') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-06::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
