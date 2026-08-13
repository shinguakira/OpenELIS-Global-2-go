-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-01-sequences::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create sequences for enrollment tables
CREATE SEQUENCE  IF NOT EXISTS clinlims.eqa_enrollment_seq START WITH 1 INCREMENT BY 1;

CREATE SEQUENCE  IF NOT EXISTS clinlims.eqa_lab_enroll_seq START WITH 1 INCREMENT BY 1;

CREATE SEQUENCE  IF NOT EXISTS clinlims.eqa_lab_unit_map_seq START WITH 1 INCREMENT BY 1;

CREATE SEQUENCE  IF NOT EXISTS clinlims.eqa_lab_test_map_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-01-sequences::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
