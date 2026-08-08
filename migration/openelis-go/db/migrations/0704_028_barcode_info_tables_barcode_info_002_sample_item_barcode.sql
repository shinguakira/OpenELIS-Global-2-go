-- source: liquibase liquibase/3.3.x.x/028-barcode-info-tables.xml::barcode-info-002-sample-item-barcode-info::ogc-284
-- +goose Up
-- +goose StatementBegin
-- Create sample_item_barcode_info table and sequence for specimen/block/slide/freezer label counts
CREATE SEQUENCE  IF NOT EXISTS clinlims.sample_item_barcode_info_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.sample_item_barcode_info (id INTEGER DEFAULT nextval('sample_item_barcode_info_seq') NOT NULL, sample_item_id numeric(10, 0) NOT NULL, print_specimen_num INTEGER, print_block_num INTEGER, print_slide_num INTEGER, print_freezer_num INTEGER, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT sample_item_barcode_info_pkey PRIMARY KEY (id), CONSTRAINT fk_sample_item_barcode_info_sample_item FOREIGN KEY (sample_item_id) REFERENCES sample_item(id), UNIQUE (sample_item_id));
CREATE INDEX IF NOT EXISTS idx_sample_item_barcode_info_sample_item_id ON clinlims.sample_item_barcode_info(sample_item_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-barcode-info-tables.xml::barcode-info-002-sample-item-barcode-info::ogc-284
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
