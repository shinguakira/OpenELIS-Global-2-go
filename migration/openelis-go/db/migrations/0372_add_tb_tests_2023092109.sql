-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092109::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.panel(id,name,description,lastupdated,sort_order,is_active,name_localization_id) VALUES
            (nextval('panel_seq'), 'MTBDRplus','MTBDRplus', now(), 51,'Y', (select id from clinlims.localization where
            french ='MTBDRplus' limit 1)),
            (nextval('panel_seq'), 'MTBDRsl','MTBDRsl', now(), 52,'Y', (select id from clinlims.localization where
            french ='MTBDRsl' limit 1)),
            (nextval('panel_seq'), 'Antibiogramme TB 1ere Ligne','Antibiogramme TB 1ere Ligne', now(), 53,'Y', (select id from clinlims.localization where
            french ='Antibiogramme TB 1ere Ligne' limit 1)),
            (nextval('panel_seq'), 'Antibiogramme TB 2e Ligne','Antibiogramme TB 2e Ligne', now(), 54,'Y', (select id from clinlims.localization where
            french ='Antibiogramme TB 2e Ligne' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092109::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
