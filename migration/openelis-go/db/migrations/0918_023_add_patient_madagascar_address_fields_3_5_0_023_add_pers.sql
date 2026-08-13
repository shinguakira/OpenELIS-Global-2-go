-- source: liquibase liquibase/3.5.x.x/023-add-patient-madagascar-address-fields.xml::3.5.0-023-add-person-fokontany::claude-uat
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.person ADD IF NOT EXISTS fokontany VARCHAR(120);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS fokontany;
-- +goose StatementEnd
