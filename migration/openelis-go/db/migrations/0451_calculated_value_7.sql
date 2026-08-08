-- source: liquibase liquibase/2.8.x.x/calculated_value.xml::7::mozzymutesa
-- +goose Up
-- +goose StatementBegin
ALTER TABLE calculation ADD CONSTRAINT calculation_name_unique_constraint UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE calculation DROP COLUMN IF EXISTS CONSTRAINT;
-- +goose StatementEnd
