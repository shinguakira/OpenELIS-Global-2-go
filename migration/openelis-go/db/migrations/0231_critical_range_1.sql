-- source: liquibase liquibase/2.7.x.x/critical_range.xml::1::cliff
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.result_limits ADD IF NOT EXISTS high_critical DOUBLE PRECISION DEFAULT 'Infinity'::double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.result_limits DROP COLUMN IF EXISTS high_critical;
-- +goose StatementEnd
