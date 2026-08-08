-- source: liquibase liquibase/2.3.x.x/new_tests.xml::331::csteele
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.test
            SET reporting_name_localization_id = (select id from localization where description = 'test report name' and english = 'HIV-1 DNA by qRT-PCR' limit 1)
            WHERE description = 'HIVINFANTVIRALLOAD(Plasma)';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::331::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
