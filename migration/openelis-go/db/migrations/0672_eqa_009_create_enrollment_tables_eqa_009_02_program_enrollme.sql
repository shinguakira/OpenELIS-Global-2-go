-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-02-program-enrollment::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create provider-side enrollment table linking organizations to EQA programs
CREATE TABLE IF NOT EXISTS clinlims.eqa_program_enrollment (id numeric(10, 0) NOT NULL, eqa_program_id numeric(10, 0) NOT NULL, organization_id numeric(10, 0) NOT NULL, enrollment_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, status VARCHAR(20) DEFAULT 'Active' NOT NULL, status_changed_date TIMESTAMP WITHOUT TIME ZONE, status_changed_by numeric(10, 0), withdrawal_reason TEXT, sys_user_id VARCHAR(20) NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT eqa_program_enrollment_pkey PRIMARY KEY (id), CONSTRAINT fk_enrollment_org FOREIGN KEY (organization_id) REFERENCES clinlims.organization(id), CONSTRAINT fk_enrollment_program FOREIGN KEY (eqa_program_id) REFERENCES clinlims.eqa_program(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.eqa_program_enrollment;
-- +goose StatementEnd
