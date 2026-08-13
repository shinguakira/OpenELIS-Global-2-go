-- source: liquibase liquibase/2.6.x.x/provider.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'restrictFreeTextProviderEntry', NOW(), 'should national ID be required on the patient entry screen', 'true', 'false', (SELECT id FROM site_information_domain WHERE name = 'sampleEntryConfig'), 'boolean', 'siteInfo.instruction.freeprovider.req.i', '0', 'siteInfo.instruction.freeprovider.req.d') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/provider.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
