-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-019-unique-nce-number::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add unique constraint on nce_number to prevent duplicate NCE numbers
ALTER TABLE clinlims.nc_event ADD CONSTRAINT uq_nc_event_nce_number UNIQUE (nce_number);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.nc_event DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
