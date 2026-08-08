-- source: liquibase liquibase/2.3.x.x/new_tests.xml::631::csteele
-- +goose Up
-- +goose StatementBegin
DELETE FROM clinlims.test_result
                WHERE tst_rslt_type = 'N' AND test_id = (select id from test where description = 'COVIDPCR(Fluid)' limit 1);
INSERT INTO clinlims.test_result(id,test_id,result_group,flags,tst_rslt_type,value,significant_digits,quant_limit,cont_level,lastupdated,scriptlet_id,sort_order,is_quantifiable,is_active,is_normal)
                VALUES
                    (nextval('test_result_seq'),
                    (select id from test where description = 'COVIDPCR(Fluid)' limit 1),
                    null,null,'D',
                    (select id from dictionary where dict_entry = 'SARS-COV-2 RNA NOT DETECTED' limit 1),
                    null,null,null,now(),null,1,'f','t','f'),
                    (nextval('test_result_seq'),
                    (select id from test where description = 'COVIDPCR(Fluid)' limit 1),
                    null,null,'D',
                    (select id from dictionary where dict_entry = 'SARS-CoV-2 RNA DETECTED' limit 1),
                    null,null,null,now(),null,1,'f','t','f'),
                    (nextval('test_result_seq'),
                    (select id from test where description = 'COVIDPCR(Fluid)' limit 1),
                    null,null,'D',
                    (select id from dictionary where dict_entry = 'RETEST - INCONCLUSIVE' limit 1),
                    null,null,null,now(),null,1,'f','t','f'),
                    (nextval('test_result_seq'),
                    (select id from test where description = 'COVIDPCR(Fluid)' limit 1),
                    null,null,'D',
                    (select id from dictionary where dict_entry = 'Invalid' limit 1),
                    null,null,null,now(),null,1,'f','t','f') ON CONFLICT DO NOTHING;
UPDATE clinlims.test t
                SET default_test_result_id = (
                    SELECT tr.id
                    FROM clinlims.test_result tr
                    WHERE CAST (tr.value as NUMERIC) in (
                        SELECT d.id FROM clinlims.dictionary d
                        WHERE d.dict_entry = 'SARS-COV-2 RNA NOT DETECTED'
                    )
                    AND tr.test_id = t.id
                )
                WHERE description = 'COVIDPCR(Fluid)';
UPDATE clinlims.result_limits rl
                SET normal_dictionary_id = (
                    SELECT d.id FROM clinlims.dictionary d
                    WHERE d.dict_entry = 'SARS-COV-2 RNA NOT DETECTED'
                )
                WHERE rl.test_id in (
                    SELECT id FROM clinlims.test
                    WHERE description = 'COVIDPCR(Fluid)'
                );
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::631::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
