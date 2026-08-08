-- source: liquibase liquibase/2.8.x.x/lab_number_alphanum_update.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.site_information SET description = 'specifies the format of the acession number,ex: SITEYEARNUM', value = 'SITEYEARNUM' WHERE name = 'acessionFormat' AND value = 'SiteYearNum';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/lab_number_alphanum_update.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
