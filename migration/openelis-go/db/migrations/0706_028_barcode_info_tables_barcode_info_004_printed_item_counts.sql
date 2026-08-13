-- source: liquibase liquibase/3.3.x.x/028-barcode-info-tables.xml::barcode-info-004-printed-item-counts::ogc-284
-- +goose Up
-- +goose StatementBegin
-- Add printed_specimen_count, printed_block_count, printed_slide_count, printed_freezer_count to sample_item_barcode_info
ALTER TABLE clinlims.sample_item_barcode_info ADD IF NOT EXISTS printed_specimen_count INTEGER DEFAULT 0;
ALTER TABLE clinlims.sample_item_barcode_info ADD IF NOT EXISTS printed_block_count INTEGER DEFAULT 0;
ALTER TABLE clinlims.sample_item_barcode_info ADD IF NOT EXISTS printed_slide_count INTEGER DEFAULT 0;
ALTER TABLE clinlims.sample_item_barcode_info ADD IF NOT EXISTS printed_freezer_count INTEGER DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-barcode-info-tables.xml::barcode-info-004-printed-item-counts::ogc-284
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
