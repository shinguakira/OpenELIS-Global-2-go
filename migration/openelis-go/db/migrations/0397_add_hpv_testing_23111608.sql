-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111608::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.result_limits(id, test_id,	test_result_type_id,min_age, max_age, gender, low_normal,high_normal, low_valid,
            high_valid, lastupdated, normal_dictionary_id,always_validate)
            VALUES (nextval('result_limits_seq'),
            (select id from	test where	description = 'HPV 16'),4,0,'Infinity',null,'-Infinity','Infinity','-Infinity','Infinity',now(),null,'f') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111608::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
