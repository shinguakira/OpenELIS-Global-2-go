-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-unique-constraint-freezer-storage-device::mherman22
-- +goose Up
-- +goose StatementBegin
-- One-to-one relationship: one StorageDevice can have at most one Freezer monitoring config
ALTER TABLE clinlims.freezer ADD CONSTRAINT uk_freezer_storage_device UNIQUE (storage_device_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.freezer DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
