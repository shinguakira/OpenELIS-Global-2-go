-- source: liquibase liquibase/2.1.x.x/lab_director_signature.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'labDirectorSignature', NOW(), 'image for lab director signature', ' ', 'false', (SELECT id FROM site_information_domain WHERE name = 'printedReportsConfig'), 'logoUpload', 'siteInfo.instruction.directorsignature.i', '0', 'siteInfo.instruction.directorsignature.d') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'labDirectorName', NOW(), 'image for lab director name', ' ', 'false', (SELECT id FROM site_information_domain WHERE name = 'printedReportsConfig'), 'text', 'siteInfo.instruction.directorname.i', '0', 'siteInfo.instruction.directorname.d') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'labDirectorTitle', NOW(), 'image for lab director title', ' ', 'false', (SELECT id FROM site_information_domain WHERE name = 'printedReportsConfig'), 'text', 'siteInfo.instruction.directortitle.i', '0', 'siteInfo.instruction.directortitle.d') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/lab_director_signature.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
