-- source: liquibase liquibase/3.2.x.x/generic_sample_menu.xml::2025110702::Generic Sample Developer
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for Generic Sample Order
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_generic_sample'), '1', 'menu_generic_sample_order', '/GenericSample/Order', 'banner.menu.generic.sample.order', 'tooltip.banner.menu.generic.sample.order', 'false', 'true') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/generic_sample_menu.xml::2025110702::Generic Sample Developer
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
