-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111702::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.panel(id,name,description,lastupdated,sort_order,is_active,name_localization_id) VALUES
            (nextval('panel_seq'), 'HPV HR','HPV HR', now(), 100,'Y', (select id from clinlims.localization where
            french ='HPV HR' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111702::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
