-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-020-add-assigned-to::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add assigned_to column to nc_event for user assignment
ALTER TABLE clinlims.nc_event ADD IF NOT EXISTS assigned_to INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.nc_event DROP COLUMN IF EXISTS assigned_to;
-- +goose StatementEnd
