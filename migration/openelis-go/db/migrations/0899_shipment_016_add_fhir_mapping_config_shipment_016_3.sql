-- source: liquibase liquibase/3.5.x.x/shipment-016-add-fhir-mapping-config.xml::shipment-016-3::oe
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, description, value, value_type, domain_id, lastupdated) VALUES (nextval('clinlims.site_information_seq'), 'fhirNonConformityCodes', 'SNOMED CT codes for non-conformities (JSON)', '{"RECEIVED_DAMAGED":"281411007","RECEIVED_LEAKED":"281412000","MISSING":"281264009","REJECTED":"123840003"}', 'text', (SELECT id FROM clinlims.site_information_domain WHERE name = 'siteIdentity'), NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-016-add-fhir-mapping-config.xml::shipment-016-3::oe
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
