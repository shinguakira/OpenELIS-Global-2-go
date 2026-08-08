-- source: liquibase liquibase/2.7.x.x/add_tb_samples.xml::2023091523::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test_section (id, name, description, is_external, lastupdated, is_active, name_localization_id, display_key) VALUES (nextval('clinlims.test_section_seq'), 'TB', 'Tuberculose', 'N', NOW(), 'Y', (SELECT id FROM clinlims.localization WHERE french ='Tuberculose' limit 1), 'test_section.TB') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_samples.xml::2023091523::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
