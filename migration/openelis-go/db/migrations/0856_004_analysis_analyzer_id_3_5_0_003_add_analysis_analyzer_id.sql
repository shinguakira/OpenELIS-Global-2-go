-- source: liquibase liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-add-analysis-analyzer-id-fk::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analysis ADD CONSTRAINT fk_analysis_analyzer FOREIGN KEY (analyzer_id) REFERENCES clinlims.analyzer (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-add-analysis-analyzer-id-fk::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
