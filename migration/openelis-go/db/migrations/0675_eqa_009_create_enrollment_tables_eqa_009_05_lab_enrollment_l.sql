-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-05-lab-enrollment-lab-unit::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create lab unit mapping table for self-enrollment
CREATE TABLE IF NOT EXISTS clinlims.eqa_lab_enrollment_lab_unit (id numeric(10, 0) NOT NULL, enrollment_id numeric(10, 0) NOT NULL, test_section_id numeric(10, 0) NOT NULL, sys_user_id VARCHAR(20) NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT eqa_lab_enrollment_lab_unit_pkey PRIMARY KEY (id), CONSTRAINT fk_lab_unit_section FOREIGN KEY (test_section_id) REFERENCES clinlims.test_section(id), CONSTRAINT fk_lab_unit_enrollment FOREIGN KEY (enrollment_id) REFERENCES clinlims.eqa_lab_program_enrollment(id));
ALTER TABLE clinlims.eqa_lab_enrollment_lab_unit ADD CONSTRAINT uq_lab_enrollment_unit UNIQUE (enrollment_id, test_section_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-05-lab-enrollment-lab-unit::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
