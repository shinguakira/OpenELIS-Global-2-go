-- source: liquibase liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-assign-role-results::pkomena
-- +goose Up
-- +goose StatementBegin
-- Assign Sample Shipment Management to Results role (read-only access)
INSERT INTO clinlims.system_role_module (id, has_select, has_add, has_update, has_delete, system_role_id, system_module_id) VALUES (nextval('clinlims.system_role_module_seq'), 'Y', 'N', 'N', 'N', (SELECT id FROM clinlims.system_role WHERE name = 'Results'), (SELECT id FROM clinlims.system_module WHERE name = 'SampleShipmentManagement')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-assign-role-results::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
