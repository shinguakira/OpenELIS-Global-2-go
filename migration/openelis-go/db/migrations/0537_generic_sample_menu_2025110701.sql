-- source: liquibase liquibase/3.2.x.x/generic_sample_menu.xml::2025110701::Generic Sample Developer
-- +goose Up
-- +goose StatementBegin
-- Add menu entry for Generic Sample as top-level menu
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), NULL, '3', 'menu_generic_sample', 'banner.menu.generic.sample', 'tooltip.banner.menu.generic.sample', 'false', 'true', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/generic_sample_menu.xml::2025110701::Generic Sample Developer
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
