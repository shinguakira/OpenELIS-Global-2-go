-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-003-create-inventory-item-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create unified inventory_item table for all inventory types
CREATE TABLE IF NOT EXISTS clinlims.inventory_item (id BIGINT NOT NULL, fhir_uuid UUID NOT NULL, name VARCHAR(255) NOT NULL, description TEXT, item_type VARCHAR(50) NOT NULL, category VARCHAR(100), manufacturer VARCHAR(255), catalog_number VARCHAR(100), storage_requirements VARCHAR(255), quantity_per_unit INTEGER, units VARCHAR(50) NOT NULL, low_stock_threshold INTEGER, expiration_alert_days INTEGER, stability_after_opening INTEGER, dilution_notes TEXT, compatible_analyzers VARCHAR(500), calibration_required VARCHAR(1) DEFAULT 'N', tests_per_kit INTEGER, individual_tracking VARCHAR(1) DEFAULT 'N', source_organization VARCHAR(255), kit_test_type VARCHAR(50), is_active VARCHAR(1) DEFAULT 'Y' NOT NULL, CONSTRAINT inventory_item_pkey PRIMARY KEY (id), UNIQUE (fhir_uuid));
ALTER TABLE clinlims.inventory_item
            ADD CONSTRAINT chk_item_type
            CHECK (item_type IN ('REAGENT', 'RDT', 'CARTRIDGE', 'HIV_KIT', 'SYPHILIS_KIT'));
ALTER TABLE clinlims.inventory_item
            ADD CONSTRAINT chk_is_active
            CHECK (is_active IN ('Y', 'N'));
ALTER TABLE clinlims.inventory_item
            ADD CONSTRAINT chk_calibration_required
            CHECK (calibration_required IN ('Y', 'N'));
ALTER TABLE clinlims.inventory_item
            ADD CONSTRAINT chk_individual_tracking
            CHECK (individual_tracking IN ('Y', 'N'));
CREATE SEQUENCE  IF NOT EXISTS clinlims.inventory_item_seq START WITH 1 INCREMENT BY 1;
CREATE INDEX IF NOT EXISTS idx_inventory_item_type ON clinlims.inventory_item(item_type);
CREATE INDEX IF NOT EXISTS idx_inventory_item_category ON clinlims.inventory_item(category);
CREATE INDEX IF NOT EXISTS idx_inventory_item_active ON clinlims.inventory_item(is_active);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-003-create-inventory-item-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
