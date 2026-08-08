-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092108::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.tb_method_test(id,method_id,test_id,is_active) VALUES
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Rifampicine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Isoniazide' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Pyrazinamide' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Ethambutol' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Capreomicine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Teridizone' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Kanamycine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Imipenem' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Streptomicine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Ethionamide' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Prothionamide' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Cyclosérine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Linezolid' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Moxifloxacine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Clofazimine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Levofloxacine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='PAS' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Imipénème-Cilastatine' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Meropenem' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Bedaquilline' limit 1),'Y'),
            (nextval('tb_method_test_seq'), (select id from dictionary where
            dict_entry ='Tests de sensibilité phénotypique' limit 1),
            (select id from test where name ='Delamanide' limit 1),'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_tests.xml::2023092108::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
