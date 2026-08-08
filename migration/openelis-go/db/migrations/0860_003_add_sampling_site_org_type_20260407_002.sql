-- source: liquibase liquibase/3.5.x.x/003-add-sampling-site-org-type.xml::20260407-002::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add 'sampling site' organization type for environmental workflows
INSERT INTO clinlims.organization_type (id, short_name, description, name_display_key) VALUES (nextval('clinlims.organization_type_seq'), 'sampling site', 'Environmental sampling site for monitoring', 'org.type.sampling.site') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-add-sampling-site-org-type.xml::20260407-002::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
