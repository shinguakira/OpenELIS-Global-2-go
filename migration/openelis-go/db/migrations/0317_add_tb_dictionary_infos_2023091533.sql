-- source: liquibase liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091533::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Diagnostic moléculaire TB',now(),'TB DiagMol',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.diagnostic.molecular',60001,null),
            (nextval('dictionary_seq'),'Y','Microscopie TB',now(),'Microsc',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.microscopy',60002,null),
            (nextval('dictionary_seq'),'Y','Culture TB',now(),'CultureTB',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.culture',60003,null),
            (nextval('dictionary_seq'),'Y','Diagnostic Immunologique TB',now(),'ImmunTB',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.immunologic_diagnostic',60004,null),
            (nextval('dictionary_seq'),'Y','Tests de sensibilité genotypique LPA',now(),'LPATests',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.lpa_tests',60005,null),
            (nextval('dictionary_seq'),'Y','Tests de sensibilité phénotypique',now(),'PhenoTests',(select id from clinlims.dictionary_category where description = 'TB Analysis Methods' limit 1),
            'dictionary.tb.phenotypic_tests',60006,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091533::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
