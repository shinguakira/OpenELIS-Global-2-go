-- source: liquibase liquibase/2.8.x.x/cytology.xml::3::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create cytology_sample table
CREATE TABLE IF NOT EXISTS cytology_sample (id INTEGER NOT NULL, technician_id numeric(10), cytopathologist_id numeric(10), specimen_adequacy_id INTEGER, program_id numeric(10), sample_id numeric(10), status VARCHAR(255), questionnaire_response_uuid UUID, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT cytology_sample_pkey PRIMARY KEY (id));
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_program_id_fk FOREIGN KEY (program_id) REFERENCES program (id);
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_sample_id_fk FOREIGN KEY (sample_id) REFERENCES sample (id);
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_technician_id_fk FOREIGN KEY (technician_id) REFERENCES system_user (id);
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_cytopathologist_id_fk FOREIGN KEY (cytopathologist_id) REFERENCES system_user (id);
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_specimen_adequacy_id_fk FOREIGN KEY (specimen_adequacy_id) REFERENCES cytology_specimen_adequacy (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::3::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
