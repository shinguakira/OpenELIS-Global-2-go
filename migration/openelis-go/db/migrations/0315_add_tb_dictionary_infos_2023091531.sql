-- source: liquibase liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091531::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Cas présumé jamais traité',now(),'never deal',(select id from clinlims.dictionary_category where description = 'TB Diagnostic Reasons' limit 1),
            'dictionary.tb.order.never_dealt',60001,null),
            (nextval('dictionary_seq'),'Y','Echec',now(),'Echec',(select id from clinlims.dictionary_category where description = 'TB Diagnostic Reasons' limit 1),
            'dictionary.tb.order.failure',60002,null),
            (nextval('dictionary_seq'),'Y','Rechute',now(),'Rechute',(select id from clinlims.dictionary_category where description = 'TB Diagnostic Reasons' limit 1),
            'dictionary.tb.order.relapse',60003,null),
            (nextval('dictionary_seq'),'Y','Reprise',now(),'Reprise',(select id from clinlims.dictionary_category where description = 'TB Diagnostic Reasons' limit 1),
            'dictionary.tb.order.resumption',60004,null),
            (nextval('dictionary_seq'),'Y','Frottis positif à M2',now(),'positifM2',(select id from clinlims.dictionary_category where description = 'TB Diagnostic Reasons' limit 1),
            'dictionary.tb.order.smear_positivem2',60005,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091531::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
