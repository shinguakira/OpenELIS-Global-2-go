-- source: liquibase liquibase/2.3.x.x/contact_tracing_fields.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'contactTracingEnabled', NOW(), 'whether fields for contact tracing should be enabled', 'false', 'false', (SELECT id FROM site_information_domain WHERE name = 'sampleEntryConfig'), 'boolean', 'siteInfo.instruction.contactTracing.i', '0', 'siteInfo.instruction.contactTracing.d') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/contact_tracing_fields.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
