-- source: liquibase liquibase/3.5.x.x/004-electronic-signature.xml::esig-012-add-site-information-toggle::openelisdev
-- +goose Up
-- +goose StatementBegin
-- Add site configuration setting to enable/disable electronic signatures
INSERT INTO clinlims.site_information (id, name, value, description, value_type, domain_id, lastupdated) VALUES (nextval('clinlims.site_information_seq'), 'electronicSignatureEnabled', 'false', 'Enable 21 CFR Part 11 compliant electronic signatures for result entry and validation workflows', 'boolean', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/004-electronic-signature.xml::esig-012-add-site-information-toggle::openelisdev
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
