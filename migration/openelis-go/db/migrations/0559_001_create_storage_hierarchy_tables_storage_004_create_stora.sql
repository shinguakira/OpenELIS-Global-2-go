-- source: liquibase liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-004-create-storage-rack-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_rack (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, label VARCHAR(100) NOT NULL, short_code VARCHAR(10), active BOOLEAN DEFAULT TRUE NOT NULL, parent_shelf_id numeric(10, 0) NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT storage_rack_pkey PRIMARY KEY (id), CONSTRAINT fk_rack_shelf FOREIGN KEY (parent_shelf_id) REFERENCES storage_shelf(id), UNIQUE (fhir_uuid));
ALTER TABLE storage_rack ADD CONSTRAINT uk_rack_label_in_shelf UNIQUE (parent_shelf_id, label);
CREATE SEQUENCE  IF NOT EXISTS storage_rack_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-004-create-storage-rack-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
