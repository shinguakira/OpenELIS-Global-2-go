-- source: liquibase liquibase/2.8.x.x/pathology.xml::6.1::csteele
-- +goose Up
-- +goose StatementBegin
-- adds in dictionary
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Haematoxylin and Eosin', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain A/Blue', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Congo Red', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Giemsa', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Gordon & Sweet', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Gram', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Grimelius', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Grocott', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Luxol fast blue', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Masson trichrome', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Masson Fontana', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain MGP', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain MSB', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain PAS', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain PAS D', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Perls', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain PTAH', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Rhodamine', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Verhoeff', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain Weigert Van G', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain ZN Lepra', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Stain ZN TB', (select id from dictionary_category where name = 'pathologist_requests' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::6.1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
