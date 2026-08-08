-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023091802::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','MTB non détecté',now(),'MTB NotD',null,
            'dictionary.tb.result.mtb_not_detected',80001,null),
            (nextval('dictionary_seq'),'Y','MTB détecté RIF sensible',now(),'MTB RifS',null,
            'dictionary.tb.result.mtb_detected_sensible',80002,null),
            (nextval('dictionary_seq'),'Y','MTB détecté RIF Résistant',now(),'MTB RifR',null,
            'dictionary.tb.result.mtb_detected_resistant',80003,null),
            (nextval('dictionary_seq'),'Y','MTB détecté RIF indéterminé',now(),'MTB RifI',null,
            'dictionary.tb.result.mtb_ind',80004,null),
            (nextval('dictionary_seq'),'Y','Erreur',now(),'Erreur',null,
            'dictionary.tb.result.error',80005,null),
            (nextval('dictionary_seq'),'Y','Invalide',now(),'Invalide',null,
            'dictionary.tb.result.invalide',80006,null),
            (nextval('dictionary_seq'),'Y','Pas de résultat',now(),'NoResult',null,
            'dictionary.tb.result.no_result',80007,null),
            (nextval('dictionary_seq'),'Y','MTB détecté',now(),'MTB DET',null,
            'dictionary.tb.result.mtb_detected',80008,null),
            (nextval('dictionary_seq'),'Y','Positif +',now(),'Pos +',null,
            'dictionary.tb.result.mtb_pos1P',80009,null),
            (nextval('dictionary_seq'),'Y','Positif ++',now(),'Pos ++',null,
            'dictionary.tb.result.mtb_pos2P',80010,null),
            (nextval('dictionary_seq'),'Y','Positif +++',now(),'Pos +++',null,
            'dictionary.tb.result.mtb_pos3P',80011,null),
            (nextval('dictionary_seq'),'Y','Positif Rare BAAR',now(),'Pos RB',null,
            'dictionary.tb.result.mtb_posRB',80012,null),
            (nextval('dictionary_seq'),'Y','Négatif',now(),'Negative',null,
            'dictionary.tb.result.negative',80013,null),
            (nextval('dictionary_seq'),'Y','Culture contaminée',now(),'Culture C',null,
            'dictionary.tb.result.culture_cont',80014,null),
            (nextval('dictionary_seq'),'Y','Culture négative',now(),'Culture N',null,
            'dictionary.tb.result.culture_neg',80015,null),
            (nextval('dictionary_seq'),'Y','Culture positive',now(),'Culture P',null,
            'dictionary.tb.result.culture_pos',80016,null),
            (nextval('dictionary_seq'),'Y','Resistant',now(),'Resistant',null,
            'dictionary.tb.result.culture_res',80017,null),
            (nextval('dictionary_seq'),'Y','Sensible',now(),'Sensible',null,
            'dictionary.tb.result.culture_sens',80018,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023091802::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
