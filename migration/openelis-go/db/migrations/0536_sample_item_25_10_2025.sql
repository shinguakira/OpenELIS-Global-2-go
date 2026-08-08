-- source: liquibase liquibase/3.2.x.x/sample_item.xml::25-10-2025::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- create sample aliquot tab menu option
INSERT INTO clinlims.menu (id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui) VALUES (nextval('clinlims.menu_seq'), '121', 'menu_aliquot', '/Aliquot', 'banner.menu.aliquot', 'banner.menu.aliquot.tooltip', TRUE, TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/sample_item.xml::25-10-2025::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
