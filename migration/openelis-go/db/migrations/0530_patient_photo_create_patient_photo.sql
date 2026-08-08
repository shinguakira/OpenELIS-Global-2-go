-- source: liquibase liquibase/3.2.x.x/patient_photo.xml::create-patient-photo::aziz_diakite
-- +goose Up
-- +goose StatementBegin
-- create patient photo table
CREATE SEQUENCE  IF NOT EXISTS clinlims.patient_photo_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS patient_photo (id INTEGER NOT NULL, patient_id VARCHAR NOT NULL, photo_data TEXT NOT NULL, thumbnail_data TEXT NOT NULL, photo_type VARCHAR(20) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT patient_photo_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/patient_photo.xml::create-patient-photo::aziz_diakite
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
