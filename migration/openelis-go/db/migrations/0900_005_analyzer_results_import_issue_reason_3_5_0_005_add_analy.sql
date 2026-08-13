-- source: liquibase liquibase/3.5.x.x/005-analyzer-results-import-issue-reason.xml::3.5.0-005-add-analyzer-results-import-issue-reason::pmanko
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer_results ADD IF NOT EXISTS import_issue_reason VARCHAR(200);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE clinlims.analyzer_results DROP COLUMN IF EXISTS import_issue_reason;
-- +goose StatementEnd
