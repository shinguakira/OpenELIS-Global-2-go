-- source: liquibase liquibase/3.5.x.x/002-eqa-enabled-config.xml::EQA-001-2026-03-17-1::mosesmutesasira
-- +goose Up
-- +goose StatementBegin
-- Add EQA enabled configuration to site_information (Order Entry Config)
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group") VALUES (nextval('clinlims.site_information_seq'), 'eqaEnabled', NOW(), 'If true, the EQA checkbox appears on Order Entry allowing a sample to be marked as an EQA sample', 'false', 'false', (SELECT id FROM site_information_domain WHERE name = 'sampleEntryConfig'), 'boolean', 'instructions.order.entry.eqa.enabled', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-eqa-enabled-config.xml::EQA-001-2026-03-17-1::mosesmutesasira
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
