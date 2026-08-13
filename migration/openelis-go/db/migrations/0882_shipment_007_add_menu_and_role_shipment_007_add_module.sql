-- source: liquibase liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-module::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add a module for Sample Shipment Management
INSERT INTO clinlims.system_module (id, name, description, has_select_flag, has_add_flag, has_update_flag, has_delete_flag) VALUES (nextval('clinlims.system_module_seq'), 'SampleShipmentManagement', 'Manage sample shipments between facilities', 'Y', 'Y', 'Y', 'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-module::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
