-- source: liquibase liquibase/3.5.x.x/shipment-008-add-capacity-column.xml::1::shipment-module
-- +goose Up
-- +goose StatementBegin
-- Add capacity column to shipping_box table
ALTER TABLE shipping_box ADD IF NOT EXISTS capacity INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shipping_box DROP COLUMN IF EXISTS capacity;
-- +goose StatementEnd
