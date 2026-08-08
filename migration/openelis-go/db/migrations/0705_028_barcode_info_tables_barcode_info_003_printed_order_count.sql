-- source: liquibase liquibase/3.3.x.x/028-barcode-info-tables.xml::barcode-info-003-printed-order-count::ogc-284
-- +goose Up
-- +goose StatementBegin
-- Add printed_order_count to sample_barcode_info for cumulative order labels printed
ALTER TABLE clinlims.sample_barcode_info ADD IF NOT EXISTS printed_order_count INTEGER DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_barcode_info DROP COLUMN IF EXISTS printed_order_count;
-- +goose StatementEnd
