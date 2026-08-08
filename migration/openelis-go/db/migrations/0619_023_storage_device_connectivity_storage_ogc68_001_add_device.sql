-- source: liquibase liquibase/3.3.x.x/023-storage-device-connectivity.xml::storage-ogc68-001-add-device-connectivity-columns::dev-team
-- +goose Up
-- +goose StatementBegin
-- Add connectivity configuration fields (IP, port, protocol) to storage_device table for network-connected equipment
ALTER TABLE storage_device ADD IF NOT EXISTS ip_address VARCHAR(45);
ALTER TABLE storage_device ADD IF NOT EXISTS port INTEGER;
ALTER TABLE storage_device ADD IF NOT EXISTS communication_protocol VARCHAR(20);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/023-storage-device-connectivity.xml::storage-ogc68-001-add-device-connectivity-columns::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
