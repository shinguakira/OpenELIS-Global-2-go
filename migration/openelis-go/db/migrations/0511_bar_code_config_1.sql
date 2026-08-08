-- source: liquibase liquibase/3.1.x.x/bar_code_config.xml::1::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- add Config to enable Selecting Bar Code Type
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Bar Code Type', NOW(), 'BarCode', 'BarCodeType') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'BARCODE', (select id from dictionary_category where name = 'BarCodeType' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'QR', (select id from dictionary_category where name = 'BarCodeType' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, dictionary_category_id, value, "group") VALUES (nextval('clinlims.site_information_seq'), 'BarCodeType', NOW(), 'Configure Bar Code Type', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'dictionary', (select id from dictionary_category where name = 'BarCodeType' limit 1), 'BARCODE', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.1.x.x/bar_code_config.xml::1::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
