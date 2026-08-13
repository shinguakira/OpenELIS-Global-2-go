-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-01-fix-file-upload-fk-type::openelis
-- +goose Up
-- +goose StatementBegin
-- Fix analyzer_file_upload.analyzer_id type (INT->NUMERIC) and set ON DELETE SET NULL for audit trail preservation
ALTER TABLE analyzer_file_upload DROP CONSTRAINT fk_analyzer_file_upload_analyzer;

ALTER TABLE analyzer_file_upload ALTER COLUMN  analyzer_id DROP NOT NULL;

ALTER TABLE analyzer_file_upload ALTER COLUMN analyzer_id TYPE numeric(10, 0) USING (analyzer_id::numeric(10, 0));

ALTER TABLE analyzer_file_upload ADD CONSTRAINT fk_analyzer_file_upload_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-01-fix-file-upload-fk-type::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
