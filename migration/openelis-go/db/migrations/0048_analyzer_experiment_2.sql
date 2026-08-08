-- source: liquibase liquibase/2.3.x.x/analyzer_experiment.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer ALTER COLUMN name TYPE VARCHAR(255) USING (name::VARCHAR(255));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/analyzer_experiment.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
