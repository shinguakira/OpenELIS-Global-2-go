-- source: liquibase liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-002-create-storage-device-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_device (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, name VARCHAR(255) NOT NULL, code VARCHAR(50) NOT NULL, type VARCHAR(20) NOT NULL, temperature_setting DECIMAL(5, 2), capacity_limit INTEGER, active BOOLEAN DEFAULT TRUE NOT NULL, parent_room_id numeric(10, 0) NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT storage_device_pkey PRIMARY KEY (id), CONSTRAINT fk_device_room FOREIGN KEY (parent_room_id) REFERENCES storage_room(id), UNIQUE (fhir_uuid));
ALTER TABLE storage_device ADD CONSTRAINT uk_device_code_in_room UNIQUE (parent_room_id, code);
ALTER TABLE clinlims.storage_device
            ADD CONSTRAINT chk_device_type
            CHECK (type IN ('freezer', 'refrigerator', 'cabinet', 'other'));
CREATE SEQUENCE  IF NOT EXISTS storage_device_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-002-create-storage-device-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
