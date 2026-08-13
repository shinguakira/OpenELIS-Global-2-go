-- source: liquibase liquibase/3.5.x.x/003-audit-trail-menu.xml::deactivate-old-audit-trail-menu-entries::systemLevelAudit
-- +goose Up
-- +goose StatementBegin
-- Deactivate old audit trail entries nested under Routine and Study
UPDATE clinlims.menu SET is_active = false
        WHERE element_id = 'menu_reports_auditTrail'
          AND parent_id != (SELECT id FROM clinlims.menu WHERE element_id = 'menu_reports');

UPDATE clinlims.menu SET is_active = false WHERE element_id = 'menu_reports_auditTrail.study';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-audit-trail-menu.xml::deactivate-old-audit-trail-menu-entries::systemLevelAudit
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
