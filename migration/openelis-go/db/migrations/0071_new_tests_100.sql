-- source: liquibase liquibase/2.3.x.x/new_tests.xml::100::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','IgM DETECTED',now(),'IgM DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.IgmDetected',54400,null),
            (nextval('dictionary_seq'),'Y','IgM NOT DETECTED',now(),'IgM NOT DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.IgmNotDetected',54410,null),

            (nextval('dictionary_seq'),'Y','IgG DETECTED',now(),'IgG DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.IggDetected',54400,null),
            (nextval('dictionary_seq'),'Y','IgG NOT DETECTED',now(),'IgG NOT DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.IggNotDetected',54410,null),

            (nextval('dictionary_seq'),'Y','DENGUE VIRUS NOT DETECTED',now(),'DENGUE VIRUW NOT DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.DengueNotDetected',54410,null),
            (nextval('dictionary_seq'),'Y','DENGUE VIRUS TYPE1 DETECTED',now(),'DENGUE VIRUW TYPE1 DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.DengueType1Detected',54420,null),
            (nextval('dictionary_seq'),'Y','DENGUE VIRUS TYPE2 DETECTED',now(),'DENGUE VIRUW TYPE2 DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.DengueType2Detected',54430,null),

            (nextval('dictionary_seq'),'Y','Inconclusive',now(),'Inconclusive',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.Inconclusive',54450,null),

            (nextval('dictionary_seq'),'Y','HIV-1 DNA NOT DETECTED',now(),'HIV-1 DNA NOT DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.Hiv1DnaNotDetected',54410,null),
            (nextval('dictionary_seq'),'Y','HIV-1 DNA DETECTED',now(),'HIV-1 DNA DETECTED',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.Hiv1DnaDetected',54420,null),
            (nextval('dictionary_seq'),'Y','Invalid',now(),'Invalid',
            (select id from dictionary_category where name = 'Haiti Lab'),
            'dictionary.result.Invalid',54430,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::100::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
