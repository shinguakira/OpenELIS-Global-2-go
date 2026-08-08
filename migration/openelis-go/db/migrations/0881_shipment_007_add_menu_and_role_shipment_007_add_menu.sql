-- source: liquibase liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-menu::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for Sample Shipment Management
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_sample'), '99', 'menu_sample_shipment', '/SampleShipment', 'banner.menu.sample.shipment', 'tooltip.banner.menu.sample.shipment', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-007-add-menu-and-role.xml::shipment-007-add-menu::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
