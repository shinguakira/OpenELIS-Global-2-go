-- source: liquibase liquibase/3.3.x.x/eqa-011-add-enrollment-program-name.xml::eqa-011-04-add-enrollment-id-to-sample-eqa::mozzy11
-- +goose Up
-- +goose StatementBegin
-- Add eqa_enrollment_id column to sample_eqa to link orders to My Programmes
ALTER TABLE clinlims.sample_eqa ADD IF NOT EXISTS eqa_enrollment_id numeric(10, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.sample_eqa DROP COLUMN IF EXISTS eqa_enrollment_id;
-- +goose StatementEnd
