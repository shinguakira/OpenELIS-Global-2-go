-- source: liquibase liquibase/3.5.x.x/005-tat-report-menu.xml::add-tat-report-menu-entry::tat-module
-- +goose Up
-- +goose StatementBegin
-- Add TAT Report as direct child of Reports menu (OGC-310)
INSERT INTO clinlims.menu (id, parent_id, presentation_order,
                                  element_id, action_url, display_key, is_active)
      SELECT nextval('clinlims.menu_seq'),
             (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports'),
             4, 'menu_reports_tatreport', '/TATReport',
             'sideNav.title.tatreport', true
      WHERE NOT EXISTS (
        SELECT 1 FROM clinlims.menu WHERE element_id = 'menu_reports_tatreport'
      )
      AND EXISTS (
        SELECT 1 FROM clinlims.menu WHERE element_id = 'menu_reports'
      ) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/005-tat-report-menu.xml::add-tat-report-menu-entry::tat-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
