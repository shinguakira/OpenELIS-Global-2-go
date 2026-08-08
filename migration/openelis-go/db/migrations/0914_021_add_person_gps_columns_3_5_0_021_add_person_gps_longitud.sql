-- source: liquibase liquibase/3.5.x.x/021-add-person-gps-columns.xml::3.5.0-021-add-person-gps-longitude::claude-uat
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.person ADD IF NOT EXISTS gps_longitude numeric(9, 6);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.person DROP COLUMN IF EXISTS gps_longitude;
-- +goose StatementEnd
