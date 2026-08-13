-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-storage-device-fk::mherman22
-- +goose Up
-- +goose StatementBegin
-- Link Freezer to StorageDevice as parent entity
ALTER TABLE clinlims.freezer ADD IF NOT EXISTS storage_device_id INTEGER;
ALTER TABLE clinlims.freezer ADD CONSTRAINT fk_freezer_storage_device FOREIGN KEY (storage_device_id) REFERENCES clinlims.storage_device (id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_freezer_storage_device ON clinlims.freezer(storage_device_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-storage-device-fk::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
