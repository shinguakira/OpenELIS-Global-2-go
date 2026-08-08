-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111601::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add entry for HPV testing
INSERT INTO clinlims.project (id, name, description, is_active, program_code, lastupdated, display_key) VALUES (nextval('clinlims.project_seq'), 'HPV Testing', 'HPV Testing', 'Y', 'HPVT', now(), 'project.HPV.name') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111601::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
