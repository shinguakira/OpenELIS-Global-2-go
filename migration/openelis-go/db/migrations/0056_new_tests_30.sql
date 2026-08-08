-- source: liquibase liquibase/2.3.x.x/new_tests.xml::30::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT into clinlims.type_of_sample
            (id, description, domain, lastupdated, local_abbrev, is_active, sort_order, name_localization_id, display_key)
            VALUES (nextval('type_of_sample_seq') , 'Respiratory Swab', 'H', now(), 'Resp Swab', 't', 0 ,
            (select id from localization where english = 'Respiratory Swab' limit 1),
            'Sample.type.Swab') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::30::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
