-- source: liquibase liquibase/2.7.x.x/site_information.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create validate accession number row in site information
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'validateAccessionNumber', NOW(), 'Validate format of accession numbers', 'true', 'false', (SELECT id FROM site_information_domain WHERE name = 'sampleEntryConfig'), 'boolean', 'instructions.accession.validate', '0', 'siteInfo.accessionValidate') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/site_information.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
