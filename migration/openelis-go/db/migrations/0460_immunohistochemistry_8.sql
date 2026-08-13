-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::8::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- add default IHC Intensity Options in Dictionary
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Strong', (select id from dictionary_category where name = 'ihc_breast_cancer_report_intensity' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'Weak', (select id from dictionary_category where name = 'ihc_breast_cancer_report_intensity' limit 1)) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, lastupdated, dict_entry, dictionary_category_id) VALUES (nextval('clinlims.dictionary_seq'), 'Y', NOW(), 'No', (select id from dictionary_category where name = 'ihc_breast_cancer_report_intensity' limit 1)) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::8::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
