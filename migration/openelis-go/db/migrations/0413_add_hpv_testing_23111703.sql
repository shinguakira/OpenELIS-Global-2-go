-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111703::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.panel_item(id,panel_id,lastupdated,test_id) VALUES
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV 16' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV 18' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='Autre HPV HR' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV P3' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV P4' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV P5' limit 1)),
            (nextval('panel_item_seq'),(select id from clinlims.panel where name ='HPV HR' limit 1),now(),(select id from clinlims.test where
            name ='HPV 18_45' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111703::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
