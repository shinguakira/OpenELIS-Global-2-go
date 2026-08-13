-- source: liquibase liquibase/2.8.x.x/menu.xml::14::vishal
-- +goose Up
-- +goose StatementBegin
-- Update pathology menu to remove dashboard submenu and add direct URL
DELETE FROM clinlims.menu WHERE element_id='menu_pathologydashboard';

UPDATE clinlims.menu SET action_url = '/PathologyDashboard' WHERE element_id='menu_pathology';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::14::vishal
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
