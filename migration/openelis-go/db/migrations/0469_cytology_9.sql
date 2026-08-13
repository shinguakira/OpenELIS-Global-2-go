-- source: liquibase liquibase/2.8.x.x/cytology.xml::9::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- create  cytology_diagnosis_result_map table
CREATE SEQUENCE  IF NOT EXISTS clinlims.cytology_diagnosis_result_map_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS cytology_diagnosis_result_map (id INTEGER NOT NULL, category VARCHAR(255), result_type VARCHAR(255), results VARCHAR(255), cytology_diagnosis_id INTEGER, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT cytology_diagnosis_result_map_pkey PRIMARY KEY (id));
ALTER TABLE cytology_diagnosis_result_map ADD CONSTRAINT cytology_diagnosis_result_map_cytology_diagnosis_id_fk FOREIGN KEY (cytology_diagnosis_id) REFERENCES cytology_diagnosis (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::9::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
