-- source: liquibase liquibase/3.5.x.x/patient_id_document.xml::create-patient-id-document::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Create patient ID document table for scanned identification cards
CREATE SEQUENCE  IF NOT EXISTS clinlims.patient_id_document_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS patient_id_document (id INTEGER NOT NULL, patient_id VARCHAR NOT NULL, document_data TEXT NOT NULL, thumbnail_data TEXT NOT NULL, document_type VARCHAR(50) NOT NULL, document_category VARCHAR(50) NOT NULL, description VARCHAR(255), deleted BOOLEAN DEFAULT FALSE NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT patient_id_document_pkey PRIMARY KEY (id));
CREATE INDEX IF NOT EXISTS idx_patient_id_document_patient_id ON patient_id_document(patient_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/patient_id_document.xml::create-patient-id-document::mozzy11
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
