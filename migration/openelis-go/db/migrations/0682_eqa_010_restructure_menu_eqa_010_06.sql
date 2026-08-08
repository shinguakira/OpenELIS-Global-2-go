-- source: liquibase liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-06::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create EQA Management parent menu item
INSERT INTO clinlims.menu (id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), '36', 'menu_eqa_mgmt', 'banner.menu.eqa.mgmt', 'banner.menu.eqa.mgmt.tooltip', FALSE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-06::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
