-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-013-add-nce-action-log-last-updated::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add last_updated column to nce_action_log table
ALTER TABLE clinlims.nce_action_log ADD IF NOT EXISTS last_updated TIMESTAMP WITHOUT TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.nce_action_log DROP COLUMN IF EXISTS last_updated;
-- +goose StatementEnd
