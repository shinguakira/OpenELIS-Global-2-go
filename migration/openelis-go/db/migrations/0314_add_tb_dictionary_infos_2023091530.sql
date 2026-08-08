-- source: liquibase liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091530::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Diagnostic',now(),'Diagn',(select id from clinlims.dictionary_category where description = 'TB Order Reasons' limit 1),
            'dictionary.tb.order.diagnostic',55000,null),
            (nextval('dictionary_seq'),'Y','Follow up',now(),'Followup',(select id from clinlims.dictionary_category where description = 'TB Order Reasons' limit 1),
            'dictionary.tb.order.followup',55000,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091530::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
