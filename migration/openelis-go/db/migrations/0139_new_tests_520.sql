-- source: liquibase liquibase/2.3.x.x/new_tests.xml::520::rossumg
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test(
                id, method_id, uom_id, description, loinc, reporting_description, sticker_req_flag, is_active, active_begin, active_end, is_reportable, time_holding, time_wait, time_ta_average, time_ta_warning, time_ta_max, label_qty, lastupdated, label_id, test_trailer_id, test_section_id, scriptlet_id, test_format_id, local_code, sort_order, name, orderable, guid, name_localization_id, reporting_name_localization_id, default_test_result_id, notify_results)
            VALUES (nextval('test_seq'),null,null,'HEPATITISCVIRALLOAD(Plasma)','94500-6',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),null,null,136,null,null,'HEPATITISCVIRALLOAD',1,'HEPATITISCVIRALLOAD','t',
            '6480988b-fd76-49bf-acae-9a1385b0f4d6',
            (select id from localization where description = 'test name' and english = 'HEPATITIS C VIRAL LOAD' limit 1),
            (select id from localization where description = 'test report name' and english = 'HCV RNA by qRT-PCR' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/new_tests.xml::520::rossumg
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
