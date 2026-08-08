-- source: liquibase liquibase/2.3.x.x/new_tests.xml::624::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','SARS-COV-2 RNA NOT DETECTED',now(),'SARS-COV-2 RNA NOT DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.sarscov2NotDetected',54510,null),
            (nextval('dictionary_seq'),'Y','SARS-CoV-2 RNA DETECTED',now(),'SARS-CoV-2 RNA DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.sarscov2Detected',54520,null),

            (nextval('dictionary_seq'),'Y','RETEST - INCONCLUSIVE',now(),'RETEST - INCONCLUSIVE',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.sarscov2Inconclusive',54530,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::624::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
