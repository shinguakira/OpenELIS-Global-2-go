-- source: liquibase liquibase/3.3.x.x/003-create-indexes.xml::storage-008-create-hierarchy-indexes::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_device_parent_room ON storage_device(parent_room_id);
CREATE INDEX IF NOT EXISTS idx_shelf_parent_device ON storage_shelf(parent_device_id);
CREATE INDEX IF NOT EXISTS idx_rack_parent_shelf ON storage_rack(parent_shelf_id);
CREATE INDEX IF NOT EXISTS idx_box_parent_rack ON storage_box(parent_rack_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/003-create-indexes.xml::storage-008-create-hierarchy-indexes::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
