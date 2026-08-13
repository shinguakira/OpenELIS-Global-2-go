-- source: liquibase liquibase/3.5.x.x/023-add-patient-madagascar-address-fields.xml::3.5.0-023-add-person-hamlet-or-lot::claude-uat
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.person ADD IF NOT EXISTS hamlet_or_lot VARCHAR(120);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS hamlet_or_lot;
-- +goose StatementEnd
