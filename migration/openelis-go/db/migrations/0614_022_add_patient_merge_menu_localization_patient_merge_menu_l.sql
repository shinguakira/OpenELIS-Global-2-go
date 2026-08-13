-- source: liquibase liquibase/3.3.x.x/022-add-patient-merge-menu-localization.xml::patient-merge-menu-localization-1::claude
-- +goose Up
-- +goose StatementBegin
-- Adds localization entries for Patient Merge menu item
INSERT INTO clinlims.localization (id, description, english, french, lastupdated)
            VALUES
            (nextval('localization_seq'), 'banner.menu.patient.merge', 'Merge Patient', 'Fusionner Patient', CURRENT_TIMESTAMP),
            (nextval('localization_seq'), 'banner.menu.patient.merge.tooltip', 'Merge duplicate patient records', 'Fusionner les dossiers patients en double', CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/022-add-patient-merge-menu-localization.xml::patient-merge-menu-localization-1::claude
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
