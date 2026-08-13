-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::2::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Add Immunohistochemistry test section
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'Immunohistochemistry test section', 'Immunohistochemistry', 'Immunohistochemistry') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.test_section (id, name, description, is_external, lastupdated, sort_order, is_active, name_localization_id, display_key) VALUES (nextval('clinlims.test_section_seq'), 'Immunohistochemistry', 'Immunohistochemistry Department', 'N', NOW(), '2147483647', 'Y', (select id from localization where description = 'Immunohistochemistry test section' and english = 'Immunohistochemistry' limit 1), 'testsection.Immunohistochemistry') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::2::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
