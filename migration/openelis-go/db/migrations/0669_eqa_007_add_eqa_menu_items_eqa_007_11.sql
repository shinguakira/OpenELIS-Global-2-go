-- source: liquibase liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-11::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Update alert_type CHECK constraint to include EQA/QC alert types
ALTER TABLE clinlims.alert DROP CONSTRAINT IF EXISTS chk_alert_type;

ALTER TABLE clinlims.alert ADD CONSTRAINT chk_alert_type
      CHECK (alert_type IN (
        'FREEZER_TEMPERATURE', 'EQUIPMENT_FAILURE', 'INVENTORY_LOW', 'SAMPLE_TRACKING', 'OTHER',
        'EQA_DEADLINE', 'SAMPLE_EXPIRATION', 'STAT_UPCOMING', 'STAT_OVERDUE', 'CRITICAL_UNACKNOWLEDGED'
      ));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-11::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
