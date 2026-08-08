-- source: liquibase liquibase/3.4.14.x/009-analyzer-discovered-source-id.xml::013-009-02::OGC-526
-- +goose Up
-- +goose StatementBegin
-- Add UNREGISTERED_SOURCE to analyzer_error.error_type CHECK constraint
ALTER TABLE clinlims.analyzer_error DROP CONSTRAINT IF EXISTS chk_error_type;

ALTER TABLE clinlims.analyzer_error
            ADD CONSTRAINT chk_error_type
            CHECK (error_type IN ('MAPPING', 'VALIDATION', 'TIMEOUT', 'PROTOCOL', 'CONNECTION', 'QC_MAPPING_INCOMPLETE', 'QC_SERVICE_UNAVAILABLE', 'UNREGISTERED_SOURCE'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/009-analyzer-discovered-source-id.xml::013-009-02::OGC-526
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
