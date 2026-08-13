-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092114::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.panel_item(id,panel_id,lastupdated,test_id) VALUES
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Capreomicine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Teridizone' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Kanamycine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Pyrazinamide' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Imipenem' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Streptomicine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Ethionamide' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Prothionamide' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Cyclosérine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Linezolid' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Moxifloxacine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Clofazimine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Levofloxacine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='PAS' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Imipénème-Cilastatine' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Meropenem' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Bedaquilline' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='Antibiogramme TB 2e Ligne' limit 1),now(),(select id from clinlims.test where
            name ='Delamanide' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092114::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
