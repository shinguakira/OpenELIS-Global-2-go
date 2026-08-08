-- source: liquibase liquibase/2.6.x.x/reporting_range.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.result_limits ADD IF NOT EXISTS low_reporting_range DOUBLE PRECISION DEFAULT '-Infinity'::double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.result_limits DROP COLUMN IF EXISTS low_reporting_range;
-- +goose StatementEnd
