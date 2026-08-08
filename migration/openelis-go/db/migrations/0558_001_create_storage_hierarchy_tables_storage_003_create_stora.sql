-- source: liquibase liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-003-create-storage-shelf-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_shelf (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, label VARCHAR(100) NOT NULL, capacity_limit INTEGER, active BOOLEAN DEFAULT TRUE NOT NULL, parent_device_id numeric(10, 0) NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT storage_shelf_pkey PRIMARY KEY (id), CONSTRAINT fk_shelf_device FOREIGN KEY (parent_device_id) REFERENCES storage_device(id), UNIQUE (fhir_uuid));
ALTER TABLE storage_shelf ADD CONSTRAINT uk_shelf_label_in_device UNIQUE (parent_device_id, label);
CREATE SEQUENCE  IF NOT EXISTS storage_shelf_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-003-create-storage-shelf-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
