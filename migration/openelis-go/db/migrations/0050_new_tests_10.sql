-- source: liquibase liquibase/2.3.x.x/new_tests.xml::10::rossumg
-- +goose Up
-- +goose StatementBegin
update localization set english = 'SARS-CoV-2 RNA by qRT-PCR', french = 'SARS-CoV-2 RNA by qRT-PCR' where description = 'test report name' and english in ( 'Covid PCR', 'Covid(PCR)');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::10::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
