-- source: liquibase liquibase/3.3.x.x/030-add-address-hierarchy-config.xml::1::openelis
-- +goose Up
-- +goose StatementBegin
-- Add configuration option for new configurable address hierarchy system
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'useNewAddressHierarchy', NOW(), 'Use new configurable address hierarchy instead of legacy Health Region/District', 'true', 'false', (SELECT id FROM site_information_domain WHERE name = 'patientEntryConfig'), 'boolean', 'instructions.address.hierarchy.new', '0', 'siteInfo.useNewAddressHierarchy') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/030-add-address-hierarchy-config.xml::1::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
