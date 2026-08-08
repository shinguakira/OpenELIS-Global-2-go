-- source: liquibase liquibase/2.3.x.x/new_tests.xml::85::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT into clinlims.sampletype_test(id, sample_type_id, test_id, is_panel)
            VALUES (nextval('sample_type_test_seq'),
            (select id from type_of_sample where description = 'Fluid' limit 1),
            (select id from test where description = 'COVIDPCR(Fluid)'),
            'f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::85::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
