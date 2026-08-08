-- source: liquibase liquibase/3.3.x.x/027-patient-merge-alter-patient-table.xml::patient-merge-002-alter-patient-table::patient-merge-backend
-- +goose Up
-- +goose StatementBegin
ALTER TABLE patient ADD IF NOT EXISTS merged_into_patient_id BIGINT;
ALTER TABLE patient ADD IF NOT EXISTS is_merged BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE patient ADD IF NOT EXISTS merge_date TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE patient ADD CONSTRAINT fk_patient_merged_into FOREIGN KEY (merged_into_patient_id) REFERENCES patient (id) ON DELETE RESTRICT;
CREATE INDEX IF NOT EXISTS idx_patient_is_merged ON patient(is_merged);
CREATE INDEX IF NOT EXISTS idx_patient_merged_into ON patient(merged_into_patient_id);
ALTER TABLE patient
            ADD CONSTRAINT chk_patient_merge_consistency
            CHECK ((is_merged = TRUE AND merged_into_patient_id IS NOT NULL AND merge_date IS NOT NULL)
                OR (is_merged = FALSE));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/027-patient-merge-alter-patient-table.xml::patient-merge-002-alter-patient-table::patient-merge-backend
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
