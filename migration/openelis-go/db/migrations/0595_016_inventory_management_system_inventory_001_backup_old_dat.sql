-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-001-backup-old-data::mherman22
-- +goose Up
-- +goose StatementBegin
-- Backup old HIV/Syphilis kit data before migration
CREATE TABLE IF NOT EXISTS clinlims.inventory_item_backup_20251207 AS
            SELECT * FROM clinlims.inventory_item;

CREATE TABLE IF NOT EXISTS clinlims.inventory_location_backup_20251207 AS
            SELECT * FROM clinlims.inventory_location;

CREATE TABLE IF NOT EXISTS clinlims.inventory_receipt_backup_20251207 AS
            SELECT * FROM clinlims.inventory_receipt;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-001-backup-old-data::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
