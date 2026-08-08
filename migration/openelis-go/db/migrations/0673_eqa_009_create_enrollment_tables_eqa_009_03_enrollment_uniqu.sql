-- source: liquibase liquibase/3.3.x.x/eqa-009-create-enrollment-tables.xml::eqa-009-03-enrollment-unique-idx::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Partial unique index to prevent duplicate active enrollments
CREATE UNIQUE INDEX IF NOT EXISTS idx_enrollment_active_unique
      ON clinlims.eqa_program_enrollment (eqa_program_id, organization_id)
      WHERE status = 'Active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_enrollment_active_unique;
-- +goose StatementEnd
