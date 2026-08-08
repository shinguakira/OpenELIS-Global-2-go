-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-003-enhance-nce-action-log::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add effectiveness review columns to nce_action_log
ALTER TABLE clinlims.nce_action_log ADD IF NOT EXISTS effective BOOLEAN;
ALTER TABLE clinlims.nce_action_log ADD IF NOT EXISTS review_comments TEXT;
ALTER TABLE clinlims.nce_action_log ADD IF NOT EXISTS reviewed_by INTEGER;
ALTER TABLE clinlims.nce_action_log ADD IF NOT EXISTS review_date date;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-003-enhance-nce-action-log::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
