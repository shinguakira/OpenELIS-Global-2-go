-- source: liquibase liquibase/3.4.x.x/012-add-organization-type-hierarchy-level.xml::add-hierarchy-level-column::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add hierarchy_level column to organization_type table for address hierarchy support
ALTER TABLE clinlims.organization_type ADD IF NOT EXISTS hierarchy_level INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.organization_type DROP COLUMN IF EXISTS hierarchy_level;
-- +goose StatementEnd
