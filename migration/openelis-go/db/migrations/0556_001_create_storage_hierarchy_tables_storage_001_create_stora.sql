-- source: liquibase liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-001-create-storage-room-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_room (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, name VARCHAR(255) NOT NULL, code VARCHAR(50) NOT NULL, description TEXT, active BOOLEAN DEFAULT TRUE NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT storage_room_pkey PRIMARY KEY (id), UNIQUE (fhir_uuid), UNIQUE (code));
CREATE SEQUENCE  IF NOT EXISTS storage_room_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-001-create-storage-room-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
