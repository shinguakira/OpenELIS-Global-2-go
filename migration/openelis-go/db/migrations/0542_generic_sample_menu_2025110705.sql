-- source: liquibase liquibase/3.2.x.x/generic_sample_menu.xml::2025110705::Generic Sample Developer
-- +goose Up
-- +goose StatementBegin
-- Add a module for Generic Sample
INSERT INTO clinlims.system_module (id, name, description, has_select_flag, has_add_flag, has_update_flag, has_delete_flag) VALUES (nextval('clinlims.system_module_seq'), 'GenericSampleView', 'Generic Sample Order Entry and Management', 'Y', 'Y', 'Y', 'N') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/generic_sample_menu.xml::2025110705::Generic Sample Developer
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
