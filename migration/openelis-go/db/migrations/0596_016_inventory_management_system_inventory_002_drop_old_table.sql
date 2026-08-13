-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-002-drop-old-tables::mherman22
-- +goose Up
-- +goose StatementBegin
-- Drop old inventory tables (backup created in previous changeset)
DROP TABLE IF EXISTS clinlims.inventory_receipt CASCADE;

DROP TABLE IF EXISTS clinlims.inventory_location CASCADE;

DROP TABLE IF EXISTS clinlims.inventory_item CASCADE;

DROP SEQUENCE IF EXISTS clinlims.inventory_item_seq;

DROP SEQUENCE IF EXISTS clinlims.inventory_location_seq;

DROP SEQUENCE IF EXISTS clinlims.inventory_receipt_seq;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-002-drop-old-tables::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
