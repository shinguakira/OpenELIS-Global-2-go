-- source: liquibase liquibase/2.3.x.x/new_tests.xml::380::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization(
                id, description, english, french, lastupdated)
                VALUES
                 (nextval('localization_seq'),'test name','HIV VIRAL LOAD','HIV VIRAL LOAD', now()),
                 (nextval('localization_seq'),'test report name','HIV-1 RNA by qRT-PCR','HIV-1 RNA by qRT-PCR', now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::380::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
