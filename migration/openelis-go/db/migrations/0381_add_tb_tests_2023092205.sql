-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092205::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.tb_method_panel(id,method_id,panel_id,is_active) VALUES
            (nextval('tb_method_panel_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from panel where name ='Antibiogramme TB 1ere Ligne' limit 1),'Y'),
            (nextval('tb_method_panel_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from panel where name ='Antibiogramme TB 2e Ligne' limit 1),'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092205::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
