-- source: liquibase liquibase/3.3.x.x/017-inventory-add-last-updated.xml::inventory-add-last-updated-002::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add last_updated column to inventory_storage_location table
ALTER TABLE clinlims.inventory_storage_location ADD IF NOT EXISTS last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.inventory_storage_location DROP COLUMN IF EXISTS last_updated;
-- +goose StatementEnd
