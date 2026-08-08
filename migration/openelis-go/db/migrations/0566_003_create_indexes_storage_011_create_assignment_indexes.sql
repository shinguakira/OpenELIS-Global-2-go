-- source: liquibase liquibase/3.3.x.x/003-create-indexes.xml::storage-011-create-assignment-indexes::sample-storage-feature
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_assignment_sample_item ON sample_storage_assignment(sample_item_id);
CREATE INDEX IF NOT EXISTS idx_assignment_location ON sample_storage_assignment(location_id, location_type);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/003-create-indexes.xml::storage-011-create-assignment-indexes::sample-storage-feature
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
