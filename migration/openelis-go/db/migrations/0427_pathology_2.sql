-- source: liquibase liquibase/2.8.x.x/pathology.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.program ADD IF NOT EXISTS questionnaire_fhir_uuid UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.program DROP COLUMN IF EXISTS questionnaire_fhir_uuid;
-- +goose StatementEnd
