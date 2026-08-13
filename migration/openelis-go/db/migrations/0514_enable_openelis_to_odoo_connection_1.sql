-- source: liquibase liquibase/3.1.x.x/enable_openelis_to_odoo_connection.xml::1::mherman22
-- +goose Up
-- +goose StatementBegin
-- add Config to enable openelis connection to odoo
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, value, "group") VALUES (nextval('clinlims.site_information_seq'), 'enableOdooConnection', NOW(), 'Enable OpenELIS to Odoo connection', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'boolean', 'false', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.1.x.x/enable_openelis_to_odoo_connection.xml::1::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
