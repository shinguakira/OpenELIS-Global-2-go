-- source: liquibase liquibase/3.2.x.x/sample-management.xml::sample-mgmt-003-add-parent-index::dev-team
-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_sample_item_parent ON sample_item(parent_sample_item_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_sample_item_parent;
-- +goose StatementEnd
