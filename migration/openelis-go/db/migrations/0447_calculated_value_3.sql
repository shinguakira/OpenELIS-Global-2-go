-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::3::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- create result_calculation table
CREATE SEQUENCE  IF NOT EXISTS clinlims.result_calculation_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS result_calculation (id INTEGER NOT NULL, calculation_id INTEGER, patient_id INTEGER, result_id INTEGER, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT result_calculation_pkey PRIMARY KEY (id));
ALTER TABLE result_calculation ADD CONSTRAINT result_calculation_calculation_id_fk FOREIGN KEY (calculation_id) REFERENCES calculation (id);
ALTER TABLE result_calculation ADD CONSTRAINT result_calculation_patient_id_fk FOREIGN KEY (patient_id) REFERENCES patient (id);
ALTER TABLE result_calculation ADD CONSTRAINT result_calculation_result_id_fk FOREIGN KEY (result_id) REFERENCES result (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/calculated_value.xml::3::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
