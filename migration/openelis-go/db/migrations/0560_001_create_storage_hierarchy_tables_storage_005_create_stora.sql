-- source: liquibase liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-005-create-storage-box-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_box (id numeric(10, 0) NOT NULL, fhir_uuid UUID NOT NULL, label VARCHAR(100) NOT NULL, type VARCHAR(50), rows INTEGER DEFAULT 0 NOT NULL, columns INTEGER DEFAULT 0 NOT NULL, position_schema_hint VARCHAR(50), short_code VARCHAR(10), active BOOLEAN DEFAULT TRUE NOT NULL, parent_rack_id numeric(10, 0) NOT NULL, sys_user_id VARCHAR(36) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT storage_box_pkey PRIMARY KEY (id), CONSTRAINT fk_box_rack FOREIGN KEY (parent_rack_id) REFERENCES storage_rack(id) ON DELETE CASCADE, UNIQUE (fhir_uuid));
ALTER TABLE storage_box ADD CONSTRAINT uk_box_label_in_rack UNIQUE (parent_rack_id, label);
ALTER TABLE clinlims.storage_box
            ADD CONSTRAINT chk_box_dimensions
            CHECK (rows >= 0 AND columns >= 0);
CREATE SEQUENCE  IF NOT EXISTS storage_box_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/001-create-storage-hierarchy-tables.xml::storage-005-create-storage-box-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
