-- source: liquibase liquibase/2.8.x.x/landing_page_config.xml::1::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- add Config to require Lab Unit At Login
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, value, "group") VALUES (nextval('clinlims.site_information_seq'), 'requireLabUnitAtLogin', NOW(), 'Require Lab Unit at Login', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'boolean', 'false', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/landing_page_config.xml::1::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
