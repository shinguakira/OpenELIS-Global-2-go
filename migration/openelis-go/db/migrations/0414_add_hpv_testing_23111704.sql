-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111704::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.sampletype_panel(id,sample_type_id,panel_id) VALUES
            (nextval('sample_type_panel_seq'),(select id from clinlims.type_of_sample where description ='Prélèvement cervico-vaginal' limit 1),
            (select id from clinlims.panel where name ='HPV HR' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111704::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
