-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23112103::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Auto Sampling',now(),'AutoSamp',(select id from clinlims.dictionary_category where description = 'HPV Sampling Method' limit 1),
            'dictionary.hpv.auto.sampling',55000,null),
            (nextval('dictionary_seq'),'Y','Health Worker Sampling',now(),'HWSamp',(select id from clinlims.dictionary_category where description = 'HPV Sampling Method' limit 1),
            'dictionary.hpv.hw.sampling',55001,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23112103::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
