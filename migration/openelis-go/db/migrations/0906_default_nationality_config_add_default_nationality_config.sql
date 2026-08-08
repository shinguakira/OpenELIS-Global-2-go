-- source: liquibase liquibase/3.5.x.x/default-nationality-config.xml::add-default-nationality-config::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Add default nationality configuration for patient entry
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'default nationality', NOW(), 'Default nationality pre-selected for new patients (must match a value from the nationality list)', '', 'false', (SELECT id FROM site_information_domain WHERE name = 'patientEntryConfig'), 'freeText', 'instructions.default.nationality', '0', 'siteInfo.defaultNationality') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/default-nationality-config.xml::add-default-nationality-config::mozzy11
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
