-- source: liquibase liquibase/3.1.x.x/provisional_clinical_diagnosis.xml::1::herbert
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.observation_history_type(id, type_name, description,lastupdated) VALUES
            (nextval('observation_history_type_seq'),'provisionalClinicalDiagnosis','Tentative diagnosis made by a clinician',now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.1.x.x/provisional_clinical_diagnosis.xml::1::herbert
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
