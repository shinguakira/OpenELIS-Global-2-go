-- source: liquibase liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::1::pkomena
-- +goose Up
-- +goose StatementBegin
-- Create box_sample_item table to replace box_sample - using SampleItem as correct granularity
CREATE SEQUENCE  IF NOT EXISTS box_sample_item_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS box_sample_item (id INTEGER NOT NULL, shipping_box_id INTEGER NOT NULL, sample_item_id INTEGER NOT NULL, added_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, position_in_box INTEGER, reception_status VARCHAR(50) DEFAULT 'PENDING', reception_notes TEXT, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT box_sample_item_pkey PRIMARY KEY (id));
ALTER TABLE box_sample_item ADD CONSTRAINT fk_box_sample_item_box FOREIGN KEY (shipping_box_id) REFERENCES shipping_box (id);
ALTER TABLE box_sample_item ADD CONSTRAINT fk_box_sample_item_sample_item FOREIGN KEY (sample_item_id) REFERENCES sample_item (id);
ALTER TABLE box_sample_item ADD CONSTRAINT uk_box_sample_item_sample_item UNIQUE (sample_item_id);
CREATE INDEX IF NOT EXISTS idx_box_sample_item_box ON box_sample_item(shipping_box_id);
CREATE INDEX IF NOT EXISTS idx_box_sample_item_sample_item ON box_sample_item(sample_item_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-012-migrate-to-sample-item.xml::1::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
