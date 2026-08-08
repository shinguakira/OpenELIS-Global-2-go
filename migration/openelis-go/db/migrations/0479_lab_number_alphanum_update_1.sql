-- source: liquibase liquibase/2.8.x.x/lab_number_alphanum_update.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- add alphanum lab number info
INSERT INTO clinlims.site_information_domain (id, name, description) VALUES (nextval('clinlims.site_information_domain_seq'), 'labNumberManagement', 'Items related to lab numbers') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, "group") VALUES (nextval('clinlims.site_information_seq'), 'alphanumAccessionPrefix', NOW(), 'accession prefix for alphanumeric lab numbers', 'false', (SELECT id FROM site_information_domain WHERE name = 'labNumberManagement'), 'text', '0') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, "group") VALUES (nextval('clinlims.site_information_seq'), 'useAlphanumAccessionPrefix', NOW(), 'accession prefix for alphanumeric lab numbers', 'false', 'false', (SELECT id FROM site_information_domain WHERE name = 'labNumberManagement'), 'boolean', '0') ON CONFLICT DO NOTHING;
UPDATE clinlims.site_information SET domain_id = (SELECT id FROM site_information_domain WHERE name = 'labNumberManagement') WHERE name = 'acessionFormat';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/lab_number_alphanum_update.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
