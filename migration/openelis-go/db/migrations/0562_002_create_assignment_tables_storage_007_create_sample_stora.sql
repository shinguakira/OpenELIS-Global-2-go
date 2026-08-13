-- source: liquibase liquibase/3.3.x.x/002-create-assignment-tables.xml::storage-007-create-sample-storage-movement-table::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sample_storage_movement (id numeric(10, 0) NOT NULL, sample_item_id numeric(10, 0) NOT NULL, previous_location_id numeric(10, 0), previous_location_type VARCHAR(20), previous_position_coordinate VARCHAR(50), new_location_id numeric(10, 0), new_location_type VARCHAR(20), new_position_coordinate VARCHAR(50), moved_by_user_id numeric(10, 0) NOT NULL, movement_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, reason TEXT, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT sample_storage_movement_pkey PRIMARY KEY (id), CONSTRAINT fk_movement_sample_item FOREIGN KEY (sample_item_id) REFERENCES sample_item(id) ON DELETE CASCADE, CONSTRAINT fk_movement_user FOREIGN KEY (moved_by_user_id) REFERENCES system_user(id));
ALTER TABLE clinlims.sample_storage_movement
            ADD CONSTRAINT chk_movement_has_location
            CHECK (
                (previous_location_id IS NOT NULL AND previous_location_type IS NOT NULL) OR
                (new_location_id IS NOT NULL AND new_location_type IS NOT NULL)
            );
ALTER TABLE clinlims.sample_storage_movement
            ADD CONSTRAINT chk_previous_location_type_valid
            CHECK (
                previous_location_type IS NULL OR
                previous_location_type IN ('device', 'shelf', 'rack', 'box')
            );
ALTER TABLE clinlims.sample_storage_movement
            ADD CONSTRAINT chk_new_location_type_valid
            CHECK (
                new_location_type IS NULL OR
                new_location_type IN ('device', 'shelf', 'rack', 'box')
            );
CREATE SEQUENCE  IF NOT EXISTS sample_storage_movement_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/002-create-assignment-tables.xml::storage-007-create-sample-storage-movement-table::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
