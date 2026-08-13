-- source: liquibase liquibase/3.3.x.x/002-create-assignment-tables.xml::storage-006-create-sample-storage-assignment-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sample_storage_assignment (id numeric(10, 0) NOT NULL, sample_item_id numeric(10, 0) NOT NULL, location_id numeric(10, 0) NOT NULL, location_type VARCHAR(20) NOT NULL, position_coordinate VARCHAR(50), assigned_by_user_id numeric(10, 0) NOT NULL, assigned_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, notes TEXT, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT sample_storage_assignment_pkey PRIMARY KEY (id), CONSTRAINT fk_assignment_sample_item FOREIGN KEY (sample_item_id) REFERENCES sample_item(id) ON DELETE CASCADE, CONSTRAINT fk_assignment_user FOREIGN KEY (assigned_by_user_id) REFERENCES system_user(id), UNIQUE (sample_item_id));
ALTER TABLE clinlims.sample_storage_assignment
            ADD CONSTRAINT chk_location_type_valid
            CHECK (location_type IN ('device', 'shelf', 'rack', 'box'));
CREATE SEQUENCE  IF NOT EXISTS sample_storage_assignment_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/002-create-assignment-tables.xml::storage-006-create-sample-storage-assignment-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
