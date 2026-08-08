-- source: liquibase liquibase/2.8.x.x/menu.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create billing tab menu option
INSERT INTO clinlims.menu (id, presentation_order, element_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), '1', 'menu_billing', '', 'banner.menu.billing', 'banner.menu.billing.tooltip', TRUE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
