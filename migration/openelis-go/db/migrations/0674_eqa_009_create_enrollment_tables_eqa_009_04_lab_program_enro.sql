-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-04-lab-program-enrollment::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Create self-enrollment table for local lab participation in external programs
CREATE TABLE IF NOT EXISTS clinlims.eqa_lab_program_enrollment (id numeric(10, 0) NOT NULL, eqa_program_id numeric(10, 0) NOT NULL, provider VARCHAR(255) NOT NULL, description TEXT, is_active BOOLEAN DEFAULT TRUE NOT NULL, created_date TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, created_by numeric(10, 0), last_modified TIMESTAMP WITHOUT TIME ZONE, sys_user_id VARCHAR(20) NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT eqa_lab_program_enrollment_pkey PRIMARY KEY (id), CONSTRAINT fk_lab_enroll_eqa_program FOREIGN KEY (eqa_program_id) REFERENCES clinlims.eqa_program(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.eqa_lab_program_enrollment;
-- +goose StatementEnd
