-- source: liquibase liquibase/2.6.x.x/address_line.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Address line 1 label', NOW(), 'label name for a place', 'Street', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', 'siteInfo.instruction.label.one', '0', 'siteInfo.label.one') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Address line 2 label', NOW(), 'label name for a place', 'Camp/Commune', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', 'siteInfo.instruction.label.two', '0', 'siteInfo.label.two') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Address line 3 label', NOW(), 'label name for a place', 'Town', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', 'siteInfo.instruction.label.three', '0', 'siteInfo.label.three') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/address_line.xml::1::cliff
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
