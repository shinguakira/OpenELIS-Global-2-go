-- source: liquibase liquibase/3.5.x.x/003-audit-trail-menu.xml::add-audit-trail-sidenav-children::systemLevelAudit
-- +goose Up
-- +goose StatementBegin
-- Add System Events and Order Events as child menu items under Audit Trail folder
-- Clear the action_url so the parent becomes a non-navigating folder
UPDATE clinlims.menu SET action_url = '' WHERE element_id = 'menu_reports_audittrail';

-- Add "System Events" child menu item
      INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, is_active)
      SELECT nextval('clinlims.menu_seq'),
             (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports_audittrail'),
             1, 'menu_reports_audittrail_system', '/AuditTrailReport?type=system',
             'sideNav.label.audittrail.systemEvents', true
      WHERE NOT EXISTS (SELECT 1 FROM clinlims.menu WHERE element_id = 'menu_reports_audittrail_system');

-- Add "Order Events" child menu item
      INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, is_active)
      SELECT nextval('clinlims.menu_seq'),
             (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports_audittrail'),
             2, 'menu_reports_audittrail_order', '/AuditTrailReport?type=order',
             'sideNav.label.audittrail.orderEvents', true
      WHERE NOT EXISTS (SELECT 1 FROM clinlims.menu WHERE element_id = 'menu_reports_audittrail_order');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-audit-trail-menu.xml::add-audit-trail-sidenav-children::systemLevelAudit
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
