-- source: liquibase liquibase/2.3.x.x/national_id_config.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'National ID required', NOW(), 'should national ID be required on the patient entry screen', 'true', 'false', (SELECT id FROM site_information_domain WHERE name = 'patientEntryConfig'), 'boolean', 'siteInfo.instruction.nationalID.req.i', '0', 'siteInfo.instruction.nationalID.req.d') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Allow duplicate national ids', NOW(), 'should national ID be allowed duplicate entries', 'true', 'false', (SELECT id FROM site_information_domain WHERE name = 'patientEntryConfig'), 'boolean', 'siteInfo.instruction.nationalID.dup.i', '0', 'siteInfo.instruction.nationalID.dup.d') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/national_id_config.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
