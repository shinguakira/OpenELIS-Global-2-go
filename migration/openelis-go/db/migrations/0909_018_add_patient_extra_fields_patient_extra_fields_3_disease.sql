-- source: liquibase liquibase/3.5.x.x/018-add-patient-extra-fields.xml::patient-extra-fields-3-disease-programme-dictionary-category::mozzy
-- +goose Up
-- +goose StatementBegin
-- Dictionary category that supplies dropdown options for the
--             Target Disease Programme field. Admins populate values via the
--             existing Dictionary management UI.
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Patient Disease Programme', NOW(), 'DISPRG', 'Patient Disease Programme') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/018-add-patient-extra-fields.xml::patient-extra-fields-3-disease-programme-dictionary-category::mozzy
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
