-- source: liquibase liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-003-seed-default-patient-report::Agaba_derrick
-- +goose Up
-- +goose StatementBegin
-- Insert default PATIENT report definition for ad-hoc patient results (demo/production)
INSERT INTO clinlims.report_definition (id, name, description, category, definition_json, created_by, created_date, last_updated, is_active, report_type, is_public) VALUES ('REP-PATIENT-DEFAULT', 'Patient Results Report', 'Default ad-hoc patient results report', 'Patient', '{"columns":[{"key":"accessionNumber","header":"Accession Number","type":"String"},{"key":"patientName","header":"Patient Name","type":"String"},{"key":"patientExternalId","header":"External ID","type":"String"},{"key":"patientGender","header":"Gender","type":"String"},{"key":"patientDateOfBirth","header":"Date of Birth","type":"String"},{"key":"organizationName","header":"Organization Name","type":"String"},{"key":"sampleCollectionDate","header":"Collection Date","type":"String"},{"key":"sampleReceivedDate","header":"Received Date","type":"String"},{"key":"clinicianName","header":"Clinician Name","type":"String"},{"key":"testName","header":"Test Name","type":"String"},{"key":"testDescription","header":"Test Description","type":"String"},{"key":"analysisStatus","header":"Status","type":"String"},{"key":"resultValue","header":"Result Value","type":"String"}]}', 'system', NOW(), NOW(), TRUE, 'PATIENT', TRUE) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-create-report-definition-table.xml::reporting-003-seed-default-patient-report::Agaba_derrick
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
