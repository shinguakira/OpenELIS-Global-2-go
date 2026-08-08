-- source: liquibase liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-02-add-menu-items-list::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Add Analyzers List sub-menu item
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), (SELECT id FROM clinlims.menu WHERE element_id = 'menu_analyzers'), '1', 'menu_analyzers_list', '/analyzers', 'analyzer.navigation.analyzersList', 'analyzer.navigation.analyzersList', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/007-insert-seed-data.xml::011-007-02-add-menu-items-list::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
