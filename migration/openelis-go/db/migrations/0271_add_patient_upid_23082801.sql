-- source: liquibase liquibase/2.7.x.x/add_patient_upid.xml::23082801::CIV developer Group
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.patient ADD IF NOT EXISTS upid_code VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.patient DROP COLUMN IF EXISTS upid_code;
-- +goose StatementEnd
