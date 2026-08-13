-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-006-create-transaction-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create transaction table for audit trail
CREATE TABLE IF NOT EXISTS clinlims.inventory_transaction (id BIGINT NOT NULL, lot_id BIGINT NOT NULL, transaction_type VARCHAR(50) NOT NULL, quantity_change DECIMAL(10, 2) NOT NULL, quantity_after DECIMAL(10, 2) NOT NULL, transaction_date TIMESTAMP WITHOUT TIME ZONE NOT NULL, reference_id BIGINT, reference_type VARCHAR(50), notes TEXT, performed_by_user INTEGER NOT NULL, CONSTRAINT inventory_transaction_pkey PRIMARY KEY (id), CONSTRAINT fk_transaction_lot FOREIGN KEY (lot_id) REFERENCES clinlims.inventory_lot(id));
ALTER TABLE clinlims.inventory_transaction
            ADD CONSTRAINT chk_transaction_type
            CHECK (transaction_type IN ('RECEIPT', 'CONSUMPTION', 'ADJUSTMENT', 'DISPOSAL', 'OPENING', 'QC_TEST', 'MANUAL'));
ALTER TABLE clinlims.inventory_transaction
            ADD CONSTRAINT chk_reference_type
            CHECK (reference_type IS NULL OR reference_type IN ('TEST_RESULT', 'RECEIPT', 'QC_RUN', 'MANUAL', 'ADJUSTMENT'));
CREATE INDEX IF NOT EXISTS idx_transaction_lot ON clinlims.inventory_transaction(lot_id);
CREATE INDEX IF NOT EXISTS idx_transaction_type ON clinlims.inventory_transaction(transaction_type);
CREATE INDEX IF NOT EXISTS idx_transaction_date ON clinlims.inventory_transaction(transaction_date);
CREATE INDEX IF NOT EXISTS idx_transaction_reference ON clinlims.inventory_transaction(reference_id, reference_type);
CREATE SEQUENCE  IF NOT EXISTS clinlims.inventory_transaction_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-006-create-transaction-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
