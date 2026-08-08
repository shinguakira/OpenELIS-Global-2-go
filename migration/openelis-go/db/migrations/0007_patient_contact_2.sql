-- source: liquibase liquibase/2.0.x.x/patient_contact.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
-- create patient contact table
CREATE TABLE IF NOT EXISTS clinlims.patient_contact (id numeric(10) NOT NULL, patient_id numeric(10), person_id numeric(60), lastupdated date, CONSTRAINT patient_contact_pkey PRIMARY KEY (id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.patient_contact_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/patient_contact.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
