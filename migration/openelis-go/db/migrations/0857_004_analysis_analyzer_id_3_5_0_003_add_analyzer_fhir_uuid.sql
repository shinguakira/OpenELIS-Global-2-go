-- source: liquibase liquibase/3.5.x.x/004-analysis-analyzer-id.xml::3.5.0-003-add-analyzer-fhir-uuid::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer ADD IF NOT EXISTS fhir_uuid UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer DROP COLUMN IF EXISTS fhir_uuid;
-- +goose StatementEnd
