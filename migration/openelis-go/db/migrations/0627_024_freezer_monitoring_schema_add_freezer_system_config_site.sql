-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-system-config-site-information::mherman22
-- +goose Up
-- +goose StatementBegin
-- Add freezer monitoring system configuration to SiteInformation (B1 remediation - replaces system_config table)
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, value, "group") VALUES (nextval('clinlims.site_information_seq'), 'freezer.modbus.tcp.port', NOW(), 'Modbus TCP port for freezer monitoring (default: 502)', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', '502', '0') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.site_information (id, name, lastupdated, description, encrypted, domain_id, value_type, value, "group") VALUES (nextval('clinlims.site_information_seq'), 'freezer.bacnet.udp.port', NOW(), 'BACnet UDP port for freezer monitoring (default: 47808)', 'false', (SELECT id FROM site_information_domain WHERE name = 'siteIdentity'), 'text', '47808', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::add-freezer-system-config-site-information::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
