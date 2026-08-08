-- source: liquibase liquibase/2.6.x.x/geographical_unit.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Geographic Unit 1 Label', NOW(), 'geographical label name for a place', 'Region', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', 'siteInfo.instruction.georaphical.one', '0', 'siteInfo.geographical.region') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'Geographic Unit 2 Label', NOW(), 'geographical label name for a place', 'District', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', 'siteInfo.instruction.georaphical.two', '0', 'siteInfo.geographical.district') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/geographical_unit.xml::1::cliff
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
