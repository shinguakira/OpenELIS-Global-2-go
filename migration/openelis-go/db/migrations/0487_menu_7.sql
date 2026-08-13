-- source: liquibase liquibase/2.8.x.x/menu.xml::7::csteele
-- +goose Up
-- +goose StatementBegin
-- insert immunochem tab menu option
INSERT INTO clinlims.menu (id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), '9', 'menu_immunochem', 'sidenav.label.immunochem', 'sidenav.label.immunochem.tooltip', FALSE, TRUE) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.menu (id, presentation_order, element_id, parent_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), '1', 'menu_immunochemdashboard', (SELECT id FROM clinlims.menu WHERE element_id = 'menu_immunochem'), '/ImmunohistochemistryDashboard', 'sidenav.label.immunochem.dashboard', 'sidenav.label.immunochem.dashboard.tooltip', FALSE, TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::7::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
