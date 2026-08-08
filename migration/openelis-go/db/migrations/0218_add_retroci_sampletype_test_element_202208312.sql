-- source: liquibase liquibase/2.6.x.x/add_retroci_sampletype_test_element.xml::202208312::CIV developer Group
-- +goose Up
-- +goose StatementBegin
-- Add elements to map retroci tests to sampletype 'Dry Tube'
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Glycémie')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Créatininémie')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Transaminases ALTL')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'p24 Ag')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Western Blot 2')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Western Blot 1')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Genie II 10')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Genie II 100')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Genie II')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Vironostika')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Murex')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Integral')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Transaminases ASTL')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Bioline')) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Dry Tube'), (SELECT id FROM clinlims.test WHERE description = 'Innolia')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_retroci_sampletype_test_element.xml::202208312::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
