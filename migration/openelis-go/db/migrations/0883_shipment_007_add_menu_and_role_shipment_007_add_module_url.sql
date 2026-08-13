-- source: liquibase liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-module-url::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add URL mapping for Sample Shipment Management module
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/SampleShipment', (SELECT id FROM clinlims.system_module WHERE name = 'SampleShipmentManagement')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-module-url::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
