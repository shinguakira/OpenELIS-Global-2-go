-- source: liquibase liquibase/2.3.x.x/new_tests.xml::80::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT into clinlims.type_of_sample( id, description, domain, lastupdated, local_abbrev, is_active, sort_order, name_localization_id, display_key)
            VALUES (nextval('type_of_sample_seq') , 'Fluid', 'H', now(), 'Fluid', 't', 0 ,
            (select id from localization where english = 'Fluid' limit 1),
            'Sample.type.Fluid') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::80::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
