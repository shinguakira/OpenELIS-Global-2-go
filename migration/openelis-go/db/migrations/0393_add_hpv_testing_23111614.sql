-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111614::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.sampletype_test (id, sample_type_id, test_id) VALUES (nextval('clinlims.sample_type_test_seq'), (SELECT id FROM clinlims.type_of_sample WHERE description = 'Prélèvement cervico-vaginal'), (SELECT id FROM clinlims.test WHERE description = 'HPV P3')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111614::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
