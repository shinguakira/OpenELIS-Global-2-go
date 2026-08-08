-- source: liquibase liquibase/3.2.x.x/result_file.xml::20250825-02-add-resultfile-to-analysis::elia
-- +goose Up
-- +goose StatementBegin
-- Creating the result_file table to store result files
ALTER TABLE analysis ADD IF NOT EXISTS result_file_id INTEGER;
ALTER TABLE analysis ADD CONSTRAINT fk_analysis_resultfile FOREIGN KEY (result_file_id) REFERENCES result_file (id) ON DELETE SET NULL;
ALTER TABLE analysis ADD CONSTRAINT uq_analysis_result_file UNIQUE (result_file_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/result_file.xml::20250825-02-add-resultfile-to-analysis::elia
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
