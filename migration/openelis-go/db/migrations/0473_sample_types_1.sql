-- source: liquibase liquibase/2.8.x.x/sample_types.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- add pathology test info
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'sampleType name', 'Tissue antemortem', 'Tissue antemortem') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.type_of_sample (id, description, lastupdated, domain, local_abbrev, is_active, display_key, name_localization_id) VALUES (nextval('clinlims.type_of_sample_seq'), 'Tissue antemortem', NOW(), 'H', 'TAM', 't', 'sample.type.anteMortemsampleType', (select id from localization where english = 'Tissue antemortem' and description = 'sampleType name' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (select id from clinlims.type_of_sample where description = 'Tissue antemortem' ), (select id from clinlims.test where description = 'Histopathology examination' )) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'sampleType name', 'Tissue post mortem', 'Tissue post mortem') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.type_of_sample (id, description, lastupdated, domain, local_abbrev, is_active, display_key, name_localization_id) VALUES (nextval('clinlims.type_of_sample_seq'), 'Tissue post mortem', NOW(), 'H', 'TAM', 't', 'sample.type.postMortemsampleType', (select id from localization where english = 'Tissue post mortem' and description = 'sampleType name' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (select id from clinlims.type_of_sample where description = 'Tissue post mortem' ), (select id from clinlims.test where description = 'Histopathology examination' )) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/sample_types.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
