-- source: liquibase liquibase/2.3.x.x/new_tests.xml::105::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization(
                id, description, english, french, lastupdated)
                VALUES
                 (nextval('localization_seq'),'test name','COVID-19 ANTIBODY IgM','COVID-19 ANITBODY IgM', now()),
                 (nextval('localization_seq'),'test report name','IgM to SARS-CoV-2 by EIA or LFA','IgM to SARS-CoV-2 by EIA or LFA', now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::105::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
