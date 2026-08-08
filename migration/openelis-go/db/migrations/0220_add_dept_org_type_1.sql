-- source: liquibase liquibase/2.7.x.x/add_dept_org_type.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- Add dept as one of the default Organization Types
INSERT INTO clinlims.organization_type (id, short_name, description, name_display_key, lastupdated) VALUES (nextval('clinlims.organization_type_seq'), 'dept', 'Organisation department', 'sample.entry.project.siteDepartmentName', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_dept_org_type.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
