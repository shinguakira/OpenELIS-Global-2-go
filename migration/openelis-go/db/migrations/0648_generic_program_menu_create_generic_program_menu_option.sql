-- source: liquibase liquibase/3.3.x.x/generic-program-menu.xml::create-generic-program-menu-option::elia
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, action_url, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_results'), '7', 'order_programmes', 'banner.menu.results.order.programmes', 'banner.menu.results.order.programmes.tooltip', FALSE, TRUE, '/genericProgram', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/generic-program-menu.xml::create-generic-program-menu-option::elia
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
