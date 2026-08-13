-- source: liquibase liquibase/2.7.x.x/add_recency_testing.xml::202309104::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.test(
            id, method_id, uom_id, description,
            loinc, reporting_description,
            sticker_req_flag, is_active,
            active_begin, active_end, is_reportable,
            time_holding, time_wait,
            time_ta_average, time_ta_warning,
            time_ta_max, label_qty, lastupdated,
            label_id, test_trailer_id,
            test_section_id, scriptlet_id,
            test_format_id, local_code,
            sort_order, name, orderable,
            guid,
            name_localization_id, reporting_name_localization_id,
            default_test_result_id, notify_results)
            VALUES
            (nextval('test_seq'),null,null,'Asante HIV-1 Rapid Recency(Plasma)','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'Asante HIV-1 Rapid Recency',1,'Asante HIV-1 Rapid Recency','t',
            '6224b40c-388d-41f5-9c70-34c69d149569',
            (select id
            from localization
            where description = 'test name' and
            english = 'Asante HIV-1 Rapid Recency' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'Rapid Test for Recent infection' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.test(
            id,
            method_id, uom_id, description, loinc, reporting_description,
            sticker_req_flag, is_active, active_begin, active_end, is_reportable,
            time_holding, time_wait, time_ta_average, time_ta_warning,
            time_ta_max, label_qty, lastupdated, label_id, test_trailer_id,
            test_section_id, scriptlet_id, test_format_id, local_code,
            sort_order, name, orderable,
            guid, name_localization_id,
            reporting_name_localization_id,
            default_test_result_id,
            notify_results)
            VALUES
            (nextval('test_seq'),null,null,'Asante HIV-1 Rapid Recency(Serum)','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'Asante HIV-1 Rapid Recency',1,'Asante HIV-1 Rapid Recency','t',
            '21d9cdc8-e58a-4ba3-9a39-52c15a0c5848',
            (select id
            from localization
            where description = 'test name' and
            english = 'Asante HIV-1 Rapid Recency' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'Rapid Test for Recent infection' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_recency_testing.xml::202309104::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
