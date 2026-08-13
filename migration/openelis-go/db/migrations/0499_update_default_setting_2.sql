-- source: liquibase liquibase/2.8.x.x/update_default_setting.xml::2::moses_mutesa
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.site_information SET description = 'Restrict Free Text Provider Entry', value = 'false' WHERE name = 'restrictFreeTextProviderEntry' AND description = 'should national ID be required on the patient entry screen';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/update_default_setting.xml::2::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
