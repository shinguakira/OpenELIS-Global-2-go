-- source: liquibase liquibase/3.3.x.x/023-storage-device-connectivity.xml::storage-ogc68-002-add-port-check-constraint::dev-team
-- +goose Up
-- +goose StatementBegin
-- Add check constraint for valid port range (1-65535)
ALTER TABLE storage_device
            ADD CONSTRAINT chk_storage_device_port_range
            CHECK (port IS NULL OR (port >= 1 AND port <= 65535));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/023-storage-device-connectivity.xml::storage-ogc68-002-add-port-check-constraint::dev-team
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
