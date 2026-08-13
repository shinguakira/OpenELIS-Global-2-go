-- source: liquibase liquibase/3.5.x.x/shipment-015-add-site-organization-uuid-config.xml::shipment-015-1::oe
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, description, value, value_type, domain_id, lastupdated) VALUES (nextval('clinlims.site_information_seq'), 'siteOrganizationFhirUuid', 'FHIR UUID of this lab''s Organization. Filters FHIR shipment imports.', '', 'text', (SELECT id FROM clinlims.site_information_domain WHERE name = 'siteIdentity'), NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-015-add-site-organization-uuid-config.xml::shipment-015-1::oe
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
