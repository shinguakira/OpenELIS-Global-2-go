-- source: liquibase liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-05-set-null-analyzer-error::openelis
-- +goose Up
-- +goose StatementBegin
-- Change analyzer_error FK to SET NULL — error logs must survive analyzer deletion for compliance
ALTER TABLE analyzer_error DROP CONSTRAINT fk_analyzer_error_analyzer;

ALTER TABLE analyzer_error ALTER COLUMN  analyzer_id DROP NOT NULL;

ALTER TABLE analyzer_error ADD CONSTRAINT fk_analyzer_error_analyzer FOREIGN KEY (analyzer_id) REFERENCES analyzer (id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.4.14.x/004-fix-analyzer-fk-types-and-cascade.xml::014-004-05-set-null-analyzer-error::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
