-- source: liquibase liquibase/2.8.x.x/pathology.xml::16::csteele
-- +goose Up
-- +goose StatementBegin
-- add pathology test info
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'test name', 'Histopathology examination', 'Histopathology examination') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'test reporting name', 'Histopathology examination', 'Histopathology examination') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.test (id, description, lastupdated, is_active, is_reportable, test_section_id, name, name_localization_id, reporting_name_localization_id, guid) VALUES (nextval('test_seq') , 'Histopathology examination', NOW(), 'Y', 'N', (select id from test_section where name = 'Pathology' limit 1), 'Histopathology examination', (select id from localization where english = 'Histopathology examination' and description = 'test name' limit 1), (select id from localization where english = 'Histopathology examination' and description = 'test reporting name' limit 1), 'fe92e08b-43ad-444e-9256-4ababc287560') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::16::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
