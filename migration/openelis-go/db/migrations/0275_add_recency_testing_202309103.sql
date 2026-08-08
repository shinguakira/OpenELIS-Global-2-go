-- source: liquibase liquibase/2.7.x.x/add_recency_testing.xml::202309103::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.dictionary(id,is_active,dict_entry,lastupdated,local_abbrev,dictionary_category_id,display_key,sort_order,name_localization_id)
            VALUES
            (nextval('dictionary_seq'),'Y','Long-Term',now(),'Long-Term',null,
            'dictionary.recency.result.long_term',55000,null),
            (nextval('dictionary_seq'),'Y','Recent',now(),'Recent',null,
            'dictionary.recency.result.recent',55000,null),
            (nextval('dictionary_seq'),'Y','Inconclusive',now(),'Inconclusive',null,
            'dictionary.recency.result.Inconclusive',55000,null),
            (nextval('dictionary_seq'),'Y','Invalid',now(),'Invalid',null,
            'dictionary.recency.result.Invalid',55000,null) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_recency_testing.xml::202309103::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
