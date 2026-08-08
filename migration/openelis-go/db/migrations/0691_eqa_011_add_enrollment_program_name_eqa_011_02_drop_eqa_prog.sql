-- source: liquibase liquibase/3.3.x.x/eqa-011-add-enrollment-program-name.xml::eqa-011-02-drop-eqa-program-fk::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Drop FK to eqa_program — My Programmes are independent of EQA Programmes
ALTER TABLE clinlims.eqa_lab_program_enrollment DROP CONSTRAINT fk_lab_enroll_eqa_program;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-011-add-enrollment-program-name.xml::eqa-011-02-drop-eqa-program-fk::mozzy11
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
