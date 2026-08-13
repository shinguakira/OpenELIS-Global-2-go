-- source: liquibase liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-005-order-label-request::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Create order_label_request table: per-order/per-sample label requests with JSONB preset snapshot (FRS §7.3)
CREATE SEQUENCE  IF NOT EXISTS clinlims.order_label_request_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.order_label_request (id INTEGER DEFAULT nextval('order_label_request_seq') NOT NULL, parent_sample_id numeric(10, 0) NOT NULL, sample_item_id numeric(10, 0), preset_id INTEGER NOT NULL, qty INTEGER NOT NULL, preset_snapshot JSONB NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT order_label_request_pkey PRIMARY KEY (id), CONSTRAINT fk_order_label_request_parent_sample FOREIGN KEY (parent_sample_id) REFERENCES sample(id), CONSTRAINT fk_order_label_request_sample_item FOREIGN KEY (sample_item_id) REFERENCES sample_item(id), CONSTRAINT fk_order_label_request_preset FOREIGN KEY (preset_id) REFERENCES label_preset(id));
ALTER TABLE clinlims.order_label_request
             ADD CONSTRAINT order_label_request_qty_nonneg CHECK (qty >= 0);
CREATE INDEX IF NOT EXISTS idx_order_label_request_parent_sample ON clinlims.order_label_request(parent_sample_id);
CREATE INDEX IF NOT EXISTS idx_order_label_request_sample_item ON clinlims.order_label_request(sample_item_id);
CREATE INDEX IF NOT EXISTS idx_order_label_request_preset ON clinlims.order_label_request(preset_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-005-order-label-request::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
