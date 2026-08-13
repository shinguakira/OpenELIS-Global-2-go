-- source: liquibase liquibase/3.5.x.x/006-storage-sidenav-rename.xml::rename-storage-samples-sidenav::locationPickerRefactor
-- +goose Up
-- +goose StatementBegin
-- Sidenav: /Storage/samples → /Storage/sample-items for the
--       new SampleItemsPage route.
UPDATE clinlims.menu
         SET action_url = '/Storage/sample-items'
       WHERE element_id = 'menu_storage_samples'
         AND action_url = '/Storage/samples';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/006-storage-sidenav-rename.xml::rename-storage-samples-sidenav::locationPickerRefactor
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
