-- source: liquibase liquibase/3.3.x.x/019-fix-assignment-location-type-check.xml::storage-019-fix-assignment-location-type-check::openelisglobal-ai-agent
-- +goose Up
-- +goose StatementBegin
-- Allow 'box' in sample_storage_assignment.location_type check constraint
ALTER TABLE clinlims.sample_storage_assignment
            DROP CONSTRAINT IF EXISTS chk_location_type_valid;

ALTER TABLE clinlims.sample_storage_assignment
            ADD CONSTRAINT chk_location_type_valid
            CHECK (
                location_type IS NULL OR
                location_type IN ('device', 'shelf', 'rack', 'box')
            );
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/019-fix-assignment-location-type-check.xml::storage-019-fix-assignment-location-type-check::openelisglobal-ai-agent
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
