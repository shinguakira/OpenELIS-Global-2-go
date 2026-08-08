-- source: liquibase liquibase/3.3.x.x/029-add-gps-coordinates-to-sample.xml::OGC-300-2026-02-11-4::mherman22
-- +goose Up
-- +goose StatementBegin
-- add GPS timeout configuration to site information
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group") VALUES (nextval('clinlims.site_information_seq'), 'gpsTimeoutSeconds', NOW(), 'Timeout in seconds for GPS location requests', '10', 'false', (SELECT id FROM site_information_domain WHERE name = 'sampleEntryConfig'), 'text', 'instructions.patient.gps.timeout', '0') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-gps-coordinates-to-sample.xml::OGC-300-2026-02-11-4::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
