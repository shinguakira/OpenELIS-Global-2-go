-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-021-add-nce-specimen-analysis-id::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add analysis_id column to nce_specimen for linking tests to NCEs
ALTER TABLE clinlims.nce_specimen ADD IF NOT EXISTS analysis_id INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.nce_specimen DROP COLUMN IF EXISTS analysis_id;
-- +goose StatementEnd
