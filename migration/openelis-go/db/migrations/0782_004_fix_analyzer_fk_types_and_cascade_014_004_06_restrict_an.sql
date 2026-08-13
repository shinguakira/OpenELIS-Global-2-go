-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-06-restrict-analyzer-results::openelis
-- +goose Up
-- +goose StatementBegin
-- Change analyzer_results FK to RESTRICT — clinical results must be explicitly handled before analyzer deletion
-- Clean up any orphan results (analyzer_id references deleted analyzers)
DELETE FROM clinlims.analyzer_results
      WHERE analyzer_id NOT IN (SELECT id FROM clinlims.analyzer);

-- Drop existing NO ACTION FK and recreate with RESTRICT
      ALTER TABLE clinlims.analyzer_results
        DROP CONSTRAINT IF EXISTS analyzer_fk;

ALTER TABLE clinlims.analyzer_results
        ADD CONSTRAINT analyzer_fk
        FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer(id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-06-restrict-analyzer-results::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
