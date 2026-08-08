-- source: liquibase liquibase/3.4.14.x/012-drop-analyzer-archive-error-directories.xml::012-drop-analyzer-archive-error-directories::pmanko
-- +goose Up
-- +goose StatementBegin
-- Drop archive_directory and error_directory columns from the analyzer
--         table. These were added in 011-complete-file-config-unification as part of
--         unifying file_import_configuration onto the analyzer entity, but the analyzer
--         bridge's file-handling flow no longer uses them — it is strictly read-only with
--         respect to watched directories and tracks all processing state in a local
--         SQLite FileStateStore instead (DIGI-UW/openelis-analyzer-bridge#34, plan
--         mellow-honking-cascade Phase 1).
-- 
--         Direct drop (no expand/contract window) because Madagascar UAT has no existing
--         analyzer rows depending on these columns, so rollback cost is acceptable.
--         If rollback is needed after this ships, restore from DB backup — there is no
--         in-code path to restore the columns without losing data.
ALTER TABLE analyzer DROP COLUMN IF EXISTS archive_directory;
ALTER TABLE analyzer DROP COLUMN IF EXISTS error_directory;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/012-drop-analyzer-archive-error-directories.xml::012-drop-analyzer-archive-error-directories::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
