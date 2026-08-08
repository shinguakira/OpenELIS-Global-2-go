-- source: liquibase liquibase/3.3.x.x/017-inventory-add-last-updated.xml::inventory-add-last-updated-001::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add last_updated column to inventory_item table for audit tracking
ALTER TABLE clinlims.inventory_item ADD IF NOT EXISTS last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.inventory_item DROP COLUMN IF EXISTS last_updated;
-- +goose StatementEnd
