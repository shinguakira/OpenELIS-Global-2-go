-- source: liquibase liquibase/3.3.x.x/003-create-indexes.xml::storage-009-create-fhir-uuid-indexes::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_room_fhir_uuid ON storage_room(fhir_uuid);
CREATE INDEX IF NOT EXISTS idx_device_fhir_uuid ON storage_device(fhir_uuid);
CREATE INDEX IF NOT EXISTS idx_shelf_fhir_uuid ON storage_shelf(fhir_uuid);
CREATE INDEX IF NOT EXISTS idx_rack_fhir_uuid ON storage_rack(fhir_uuid);
CREATE INDEX IF NOT EXISTS idx_box_fhir_uuid ON storage_box(fhir_uuid);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/003-create-indexes.xml::storage-009-create-fhir-uuid-indexes::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
