-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092113::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.panel_item(id,panel_id,lastupdated,test_id) VALUES
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 1ere Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Rifampicine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 1ere Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Isoniazide' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 1ere Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Pyrazinamide' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 1ere Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Ethambutol' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092113::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
