-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-unique-constraint-freezer-storage-device::mherman22
-- +goose Up
-- +goose StatementBegin
-- One-to-one relationship: one StorageDevice can have at most one Freezer monitoring config
ALTER TABLE clinlims.freezer ADD CONSTRAINT uk_freezer_storage_device UNIQUE (storage_device_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-unique-constraint-freezer-storage-device::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
