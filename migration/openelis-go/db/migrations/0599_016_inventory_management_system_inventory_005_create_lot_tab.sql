-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-005-create-lot-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create inventory lot table for tracking specific batches
CREATE TABLE IF NOT EXISTS clinlims.inventory_lot (id BIGINT NOT NULL, fhir_uuid UUID NOT NULL, inventory_item_id BIGINT NOT NULL, storage_location_id BIGINT, lot_number VARCHAR(100) NOT NULL, expiration_date TIMESTAMP WITHOUT TIME ZONE, date_opened TIMESTAMP WITHOUT TIME ZONE, calculated_expiry_after_opening TIMESTAMP WITHOUT TIME ZONE, receipt_date TIMESTAMP WITHOUT TIME ZONE, initial_quantity DECIMAL(10, 2) NOT NULL, current_quantity DECIMAL(10, 2) NOT NULL, qc_status VARCHAR(20) DEFAULT 'PENDING' NOT NULL, status VARCHAR(20) DEFAULT 'ACTIVE' NOT NULL, barcode VARCHAR(100), version INTEGER DEFAULT 0 NOT NULL, CONSTRAINT inventory_lot_pkey PRIMARY KEY (id), CONSTRAINT fk_lot_storage_location FOREIGN KEY (storage_location_id) REFERENCES clinlims.inventory_storage_location(id), CONSTRAINT fk_lot_inventory_item FOREIGN KEY (inventory_item_id) REFERENCES clinlims.inventory_item(id), UNIQUE (fhir_uuid), UNIQUE (barcode));
ALTER TABLE clinlims.inventory_lot
            ADD CONSTRAINT chk_lot_qc_status
            CHECK (qc_status IN ('PENDING', 'PASSED', 'FAILED', 'QUARANTINED'));
ALTER TABLE clinlims.inventory_lot
            ADD CONSTRAINT chk_lot_status
            CHECK (status IN ('ACTIVE', 'IN_USE', 'EXPIRED', 'CONSUMED', 'DISPOSED', 'QUARANTINED'));
ALTER TABLE clinlims.inventory_lot
            ADD CONSTRAINT chk_lot_quantities
            CHECK (initial_quantity >= 1 AND current_quantity >= 0);
CREATE INDEX IF NOT EXISTS idx_lot_inventory_item ON clinlims.inventory_lot(inventory_item_id);
CREATE INDEX IF NOT EXISTS idx_lot_storage_location ON clinlims.inventory_lot(storage_location_id);
CREATE INDEX IF NOT EXISTS idx_lot_status ON clinlims.inventory_lot(status);
CREATE INDEX IF NOT EXISTS idx_lot_qc_status ON clinlims.inventory_lot(qc_status);
CREATE INDEX IF NOT EXISTS idx_lot_expiration ON clinlims.inventory_lot(expiration_date);
CREATE INDEX IF NOT EXISTS idx_lot_number ON clinlims.inventory_lot(lot_number);
CREATE SEQUENCE  IF NOT EXISTS clinlims.inventory_lot_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-005-create-lot-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
