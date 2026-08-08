-- source: liquibase liquibase/2.3.x.x/new_tests.xml::185::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test_result(id,test_id,result_group,flags,tst_rslt_type,value,significant_digits,quant_limit,cont_level,lastupdated,scriptlet_id,sort_order,is_quantifiable,is_active,is_normal)
            VALUES
        (nextval('test_result_seq'),
        (select id from test where description = 'COVID-19ANTIBODYIgG(Serum)'),
        null,null,'D',
        (select id from dictionary where dict_entry = 'IgG DETECTED' limit 1),
        null,null,null,now(),null,1,'f','t','f'),
        (nextval('test_result_seq'),
        (select id from test where description = 'COVID-19ANTIBODYIgG(Serum)'),
        null,null,'D',
        (select id from dictionary where dict_entry = 'IgG NOT DETECTED' limit 1),
        null,null,null,now(),null,1,'f','t','f'),
        (nextval('test_result_seq'),
        (select id from test where description = 'COVID-19ANTIBODYIgG(Serum)'),
        null,null,'D',
        (select id from dictionary where dict_entry = 'Invalid' limit 1),
        null,null,null,now(),null,1,'f','t','f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::185::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
