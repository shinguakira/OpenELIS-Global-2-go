-- source: liquibase liquibase/2.6.x.x/add_menu_study_electronic_order.xml::202208292::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add a module for study electronic orders
INSERT INTO clinlims.system_module (id, name, description, has_select_flag, has_add_flag, has_update_flag, has_delete_flag) VALUES (nextval('clinlims.system_module_seq'), 'StudyElectronicOrderView', '', 'Y', 'Y', 'Y', 'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_menu_study_electronic_order.xml::202208292::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
