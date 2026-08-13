-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-007-create-usage-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create usage table for linking inventory to test results
CREATE TABLE IF NOT EXISTS clinlims.inventory_usage (id BIGINT NOT NULL, inventory_item_id BIGINT NOT NULL, lot_id BIGINT NOT NULL, test_result_id BIGINT, analysis_id BIGINT, quantity_used DECIMAL(10, 2) NOT NULL, usage_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, performed_by_user INTEGER NOT NULL, CONSTRAINT inventory_usage_pkey PRIMARY KEY (id), CONSTRAINT fk_usage_lot FOREIGN KEY (lot_id) REFERENCES clinlims.inventory_lot(id), CONSTRAINT fk_usage_inventory_item FOREIGN KEY (inventory_item_id) REFERENCES clinlims.inventory_item(id));
ALTER TABLE clinlims.inventory_usage
            ADD CONSTRAINT chk_usage_quantity
            CHECK (quantity_used >= 1);
CREATE INDEX IF NOT EXISTS idx_usage_inventory_item ON clinlims.inventory_usage(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_usage_lot ON clinlims.inventory_usage(lot_id);
CREATE INDEX IF NOT EXISTS idx_usage_date ON clinlims.inventory_usage(usage_date);
CREATE INDEX IF NOT EXISTS idx_usage_test_result ON clinlims.inventory_usage(test_result_id);
CREATE INDEX IF NOT EXISTS idx_usage_analysis ON clinlims.inventory_usage(analysis_id);
CREATE SEQUENCE  IF NOT EXISTS clinlims.inventory_usage_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-007-create-usage-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
