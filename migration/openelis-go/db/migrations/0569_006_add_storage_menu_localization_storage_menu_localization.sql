-- source: liquibase liquibase/3.3.x.x/006-add-storage-menu-localization.xml::storage-menu-localization-1::pmanko
-- +goose Up
-- +goose StatementBegin
-- Adds localization entries for Storage menu item
INSERT INTO clinlims.localization (id, description, english, french, lastupdated)
            VALUES
            (nextval('localization_seq'), 'banner.menu.storage', 'Storage', 'Stockage', CURRENT_TIMESTAMP),
            (nextval('localization_seq'), 'banner.menu.storage.tooltip', 'Manage sample storage locations', 'Gérer les emplacements de stockage des échantillons', CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/006-add-storage-menu-localization.xml::storage-menu-localization-1::pmanko
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
