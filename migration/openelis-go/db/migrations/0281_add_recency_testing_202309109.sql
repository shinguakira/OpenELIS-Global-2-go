-- source: liquibase liquibase/2.7.x.x/add_recency_testing.xml::202309109::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO
            clinlims.test_result(id,test_id,result_group,flags,tst_rslt_type,value,significant_digits,quant_limit,cont_level,lastupdated,scriptlet_id,sort_order,is_quantifiable,is_active,is_normal)
            VALUES
            (nextval('test_result_seq'),
            (select id from test where description = 'Asante HIV-1 Rapid Recency(Serum)'),	null,null,'D',
            (select id from dictionary where dict_entry = 'Long-Term' limit 1),
            null,null,null,now(),null,1,'f','t','f'),
            (nextval('test_result_seq'),
            (select id from test where description = 'Asante HIV-1 Rapid Recency(Serum)'),
            null,null,'D',
            (select id from dictionary where dict_entry = 'Recent' limit 1),
            null,null,null,now(),null,2,'f','t','f'),
            (nextval('test_result_seq'),
            (select id from test where description ='Asante HIV-1 Rapid Recency(Serum)'),
            null,null,'D',(select id from dictionary where dict_entry = 'Inconclusive' limit 1),
            null,null,null,now(),null,3,'f','t','f'),
            (nextval('test_result_seq'),(select id from test where description ='Asante HIV-1 Rapid Recency(Serum)'),
            null,null,'D',
            (select id from dictionary where dict_entry = 'Invalid' limit 1),null,null,null,now(),null,4,'f','t','f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_recency_testing.xml::202309109::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
