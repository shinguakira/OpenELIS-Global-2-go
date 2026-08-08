-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111604::CIV developer Group
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
            (nextval('test_seq'),null,null,'HPV 16','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV 16',1,'HPV 16','t',
            '7032dd53-47aa-4b64-9349-b781675720b3',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV 16' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV 16' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'HPV 18','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV 18',1,'HPV 18','t',
            '69021779-30e0-48db-be22-006d48d83a28',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV 18' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV 18' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'Autre HPV HR','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'Autre HPV HR',1,'Autre HPV HR','t',
            'df40ca99-f4ce-44a0-9359-4e0fc591f4ca',
            (select id
            from localization
            where description = 'test name' and
            english = 'Other HR HPV' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'Other HR HPV' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'HPV 18_45','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV 18_45',1,'HPV 18_45','t',
            '7461c494-4ad9-45b8-94ed-026a2cae2e17',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV 18_45' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV 18_45' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'HPV P3','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV P3',1,'HPV P3','t',
            'f244ea85-bf64-480f-92cc-a777b408cae6',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV P3' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV P3' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'HPV P4','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV P4',1,'HPV P4','t',
            '190f3db5-f881-4b3b-af5a-1a9c3ee64e43',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV P4' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV P4' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
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
            (nextval('test_seq'),null,null,'HPV P5','',null,null,'Y',null,null,'N',null,null,null,null,null,null,now(),
            null,null,136,null,null,'HPV P5',1,'HPV P5','t',
            'e1c6d0dd-92c1-4921-8cdb-afab4cb4c894',
            (select id
            from localization
            where description = 'test name' and
            english = 'HPV P5' limit
            1),
            (select id from localization where
            description = 'test report name' and english = 'HPV P5' limit 1),
            null,'f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111604::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
