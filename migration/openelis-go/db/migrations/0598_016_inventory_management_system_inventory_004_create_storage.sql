-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-004-create-storage-location-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create hierarchical storage location table
CREATE TABLE IF NOT EXISTS clinlims.inventory_storage_location (id BIGINT NOT NULL, fhir_uuid UUID NOT NULL, name VARCHAR(255) NOT NULL, location_code VARCHAR(50), location_type VARCHAR(50) NOT NULL, description TEXT, temperature_min DECIMAL(5, 2), temperature_max DECIMAL(5, 2), parent_location_id BIGINT, is_active BOOLEAN DEFAULT TRUE NOT NULL, CONSTRAINT inventory_storage_location_pkey PRIMARY KEY (id), CONSTRAINT fk_storage_location_parent FOREIGN KEY (parent_location_id) REFERENCES clinlims.inventory_storage_location(id), UNIQUE (fhir_uuid), UNIQUE (location_code));
ALTER TABLE clinlims.inventory_storage_location
            ADD CONSTRAINT chk_location_type
            CHECK (location_type IN ('ROOM', 'REFRIGERATOR', 'FREEZER', 'SHELF', 'DRAWER', 'CABINET'));
CREATE INDEX IF NOT EXISTS idx_storage_location_type ON clinlims.inventory_storage_location(location_type);
CREATE INDEX IF NOT EXISTS idx_storage_location_parent ON clinlims.inventory_storage_location(parent_location_id);
CREATE INDEX IF NOT EXISTS idx_storage_location_active ON clinlims.inventory_storage_location(is_active);
CREATE SEQUENCE  IF NOT EXISTS clinlims.inventory_storage_location_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-004-create-storage-location-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
