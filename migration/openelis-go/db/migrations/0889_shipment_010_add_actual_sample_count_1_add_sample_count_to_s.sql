-- source: liquibase liquibase/3.5.x.x/shipment-010-add-actual-sample-count.xml::1-add-sample-count-to-shipping-box::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add actual_sample_count column to shipping_box table
ALTER TABLE shipping_box ADD IF NOT EXISTS actual_sample_count INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE shipping_box DROP COLUMN IF EXISTS actual_sample_count;
-- +goose StatementEnd
