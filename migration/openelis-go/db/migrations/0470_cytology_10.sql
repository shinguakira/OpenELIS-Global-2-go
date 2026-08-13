-- source: liquibase liquibase/2.8.x.x/cytology.xml::10::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- add columns to the cytology_sample table
ALTER TABLE clinlims.cytology_sample ADD IF NOT EXISTS cytology_diagnosis_id INTEGER;
ALTER TABLE cytology_sample ADD CONSTRAINT cytology_sample_cytology_diagnosis_id_fk FOREIGN KEY (cytology_diagnosis_id) REFERENCES cytology_diagnosis (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::10::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
