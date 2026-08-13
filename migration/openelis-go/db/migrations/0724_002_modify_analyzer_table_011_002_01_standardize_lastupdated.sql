-- source: liquibase liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-01-standardize-lastupdated-columns::madagascar-analyzer-integration
-- +goose Up
-- +goose StatementBegin
-- Rename lastupdated to last_updated in legacy analyzer tables to match BaseObject convention
ALTER TABLE analyzer RENAME COLUMN lastupdated TO last_updated;

ALTER TABLE analyzer_test_map RENAME COLUMN lastupdated TO last_updated;

ALTER TABLE ANALYZER_RESULTS RENAME COLUMN LASTUPDATED TO last_updated;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.x.x/002-modify-analyzer-table.xml::011-002-01-standardize-lastupdated-columns::madagascar-analyzer-integration
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
