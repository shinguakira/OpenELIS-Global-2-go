-- source: liquibase liquibase/2.8.x.x/pathology.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
-- create pathology_sample table
CREATE SEQUENCE  IF NOT EXISTS clinlims.program_sample_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS program_sample (id INTEGER NOT NULL, program_id numeric(10), sample_id numeric(10), questionnaire_response_uuid UUID, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT program_sample_pkey PRIMARY KEY (id));
ALTER TABLE program_sample ADD CONSTRAINT program_sample_program_id_fk FOREIGN KEY (program_id) REFERENCES program (id);
ALTER TABLE program_sample ADD CONSTRAINT program_sample_sample_id_fk FOREIGN KEY (sample_id) REFERENCES sample (id);
CREATE TABLE IF NOT EXISTS pathology_sample (id INTEGER NOT NULL, technician_id numeric(10), pathologist_id numeric(10), program_id numeric(10), sample_id numeric(10), status VARCHAR(255), questionnaire_response_uuid UUID, gross_exam VARCHAR(255), microscopy_exam VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pathology_sample_pkey PRIMARY KEY (id));
ALTER TABLE pathology_sample ADD CONSTRAINT pathology_sample_program_id_fk FOREIGN KEY (program_id) REFERENCES program (id);
ALTER TABLE pathology_sample ADD CONSTRAINT pathology_sample_sample_id_fk FOREIGN KEY (sample_id) REFERENCES sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
