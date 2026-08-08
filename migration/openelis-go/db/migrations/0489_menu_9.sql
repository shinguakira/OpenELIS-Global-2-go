-- source: liquibase liquibase/2.8.x.x/menu.xml::9::csteele
-- +goose Up
-- +goose StatementBegin
-- update results search menu options
INSERT INTO clinlims.menu (id, presentation_order, element_id, parent_id, action_url, display_key, tool_tip_key, new_window, is_active) VALUES (nextval('clinlims.menu_seq'), '6', 'menu_results_search_testdate', (SELECT id FROM clinlims.menu WHERE element_id = 'menu_results'), '/result?type=date&doRange=false', 'sidenav.label.results.testdate', 'sidenav.label.results.testdate.tooltip', FALSE, TRUE) ON CONFLICT DO NOTHING;
UPDATE clinlims.menu SET parent_id = (SELECT id FROM clinlims.menu WHERE element_id = 'menu_results') WHERE parent_id=(SELECT id FROM clinlims.menu WHERE element_id = 'menu_results_search');
DELETE FROM clinlims.menu WHERE element_id='menu_results_search';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::9::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
