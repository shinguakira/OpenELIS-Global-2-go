-- source: liquibase liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091532::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Examen de suivi 1ère ligne (TB Sensible)',now(),'TB Line1',(select id from clinlims.dictionary_category where description = 'TB Followup Reasons' limit 1),
            'dictionary.tb.followup.line1',60001,null),
            (nextval('dictionary_seq'),'Y','Examen de suivi 2ième ligne (TB/RR ; TBXDR)',now(),'TB Line2',(select id from clinlims.dictionary_category where description = 'TB Followup Reasons' limit 1),
            'dictionary.tb.followup.line2',60002,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091532::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
