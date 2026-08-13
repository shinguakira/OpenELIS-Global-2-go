-- source: liquibase liquibase/3.5.x.x/003-audit-trail-menu.xml::add-audit-trail-menu-entry::systemLevelAudit
-- +goose Up
-- +goose StatementBegin
-- Add Audit Trail as top-level Reports sub-menu (sibling of Routine/Study)
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, is_active)
      SELECT nextval('clinlims.menu_seq'),
             (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports'),
             3, 'menu_reports_audittrail', '/AuditTrailReports', 'sideNav.title.audittrail', true
      WHERE NOT EXISTS (SELECT 1 FROM clinlims.menu WHERE element_id = 'menu_reports_audittrail') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-audit-trail-menu.xml::add-audit-trail-menu-entry::systemLevelAudit
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
