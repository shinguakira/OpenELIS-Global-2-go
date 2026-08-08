-- source: liquibase liquibase/3.3.x.x/029-add-storage-subnav-items.xml::update-freezer-monitoring-no-action::navbar-extraction
-- +goose Up
-- +goose StatementBegin
-- Remove action_url from Cold Storage Monitoring to make it expandable with child items
UPDATE clinlims.menu SET action_url = '' WHERE element_id = 'menu_freezer_monitoring';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-add-storage-subnav-items.xml::update-freezer-monitoring-no-action::navbar-extraction
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
