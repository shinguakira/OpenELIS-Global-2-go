-- source: liquibase liquibase/2.8.x.x/pathology.xml::12::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'pathology test section', 'Pathology', 'Pathology') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.test_section (id, name, description, is_external, lastupdated, sort_order, is_active, name_localization_id, display_key) VALUES (nextval('clinlims.test_section_seq'), 'Pathology', 'Pathology Department', 'N', NOW(), '2147483647', 'Y', (select id from localization where description = 'pathology test section' and english = 'Pathology' limit 1), 'testsection.Pathology') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::12::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
