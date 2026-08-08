-- source: liquibase liquibase/2.6.x.x/localization.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'EIA', 'EIE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'PCR', 'réaction en chaîne par polymérase') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'STAIN', 'TACHE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'CULTURE', 'CULTURE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'PROBE', 'SONDE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'BIOCHEMICAL', 'BIOCHIMIQUE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'HIV_TEST_KIT', 'HIV_TEST_KIT') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'Diane Test', 'Diane Test') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'HPLC', 'CLHP') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'DNA SEQUENCING', 'ADN SÉQUENÇAGE') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'AUTO', 'AUTO') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'MANUAL', 'MANUEL') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization (id, lastupdated, description, english, french) VALUES (nextval('clinlims.localization_seq'), NOW(), 'The method used for tests', 'SYPHILIS_TEST_KIT', 'SYPHILIS_TEST_KIT') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/localization.xml::1::cliff
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
