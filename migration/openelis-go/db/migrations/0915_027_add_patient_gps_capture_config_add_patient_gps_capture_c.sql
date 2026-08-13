-- source: liquibase liquibase/3.5.x.x/027-add-patient-gps-capture-config.xml::add-patient-gps-capture-config::openelis
-- +goose Up
-- +goose StatementBegin
-- Admin toggle for patient registration GPS lat/long capture . Default off in core OE2; configurable under Admin > General Configuration > Patient Configuration Menu.
INSERT INTO clinlims.site_information (id, name, lastupdated, description, value, encrypted, domain_id, value_type, instruction_key, "group", description_key) VALUES (nextval('clinlims.site_information_seq'), 'patientGpsCaptureEnabled', NOW(), 'Enable GPS latitude/longitude capture on patient registration', 'false', 'false', (SELECT id FROM site_information_domain WHERE name = 'patientEntryConfig'), 'boolean', 'instructions.patient.gps.capture', '0', 'siteInfo.patientGpsCaptureEnabled') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/027-add-patient-gps-capture-config.xml::add-patient-gps-capture-config::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
