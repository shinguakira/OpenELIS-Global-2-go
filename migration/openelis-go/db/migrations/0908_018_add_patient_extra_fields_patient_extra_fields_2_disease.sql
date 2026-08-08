-- source: liquibase liquibase/3.5.x.x/018-add-patient-extra-fields.xml::patient-extra-fields-2-disease-programme-identity-type::mozzy
-- +goose Up
-- +goose StatementBegin
-- Identity type for the patient's target disease programme.
INSERT INTO clinlims.patient_identity_type (id, identity_type, description, lastupdated) VALUES (nextval('clinlims.patient_identity_type_seq'), 'DISEASE_PROGRAMME', 'Target Disease Programme', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/018-add-patient-extra-fields.xml::patient-extra-fields-2-disease-programme-identity-type::mozzy
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
