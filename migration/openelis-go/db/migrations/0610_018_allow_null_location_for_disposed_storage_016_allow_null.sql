-- source: liquibase liquibase/3.3.x.x/018-allow-null-location-for-disposed.xml::storage-016-allow-null-location-for-disposed::ogc-144
-- +goose Up
-- +goose StatementBegin
-- Allow NULL location_id and location_type in sample_storage_assignment
--             to support disposal workflow (FR-056, FR-057). Disposed samples keep
--             their assignment record with NULL location for metrics counting and
--             audit trail.
ALTER TABLE sample_storage_assignment ALTER COLUMN  location_id DROP NOT NULL;

ALTER TABLE sample_storage_assignment ALTER COLUMN  location_type DROP NOT NULL;

ALTER TABLE clinlims.sample_storage_assignment
            DROP CONSTRAINT IF EXISTS chk_location_type_valid;

ALTER TABLE clinlims.sample_storage_assignment
            ADD CONSTRAINT chk_location_type_valid
            CHECK (location_type IS NULL OR location_type IN ('device', 'shelf', 'rack'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/018-allow-null-location-for-disposed.xml::storage-016-allow-null-location-for-disposed::ogc-144
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
