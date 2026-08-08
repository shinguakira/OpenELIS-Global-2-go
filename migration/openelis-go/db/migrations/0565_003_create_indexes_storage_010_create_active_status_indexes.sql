-- source: liquibase liquibase/3.3.x.x/003-create-indexes.xml::storage-010-create-active-status-indexes::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_room_active ON storage_room(active);
CREATE INDEX IF NOT EXISTS idx_device_active ON storage_device(active);
CREATE INDEX IF NOT EXISTS idx_shelf_active ON storage_shelf(active);
CREATE INDEX IF NOT EXISTS idx_rack_active ON storage_rack(active);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/003-create-indexes.xml::storage-010-create-active-status-indexes::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
