-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-011-add-nce-specimen-last-updated::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add last_updated column to nce_specimen table
ALTER TABLE clinlims.nce_specimen ADD IF NOT EXISTS last_updated TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.nce_specimen DROP COLUMN IF EXISTS last_updated;
-- +goose StatementEnd
