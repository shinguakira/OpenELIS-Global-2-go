-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::7::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- adds Breast Cancer Hormone Receptor Status Report Parameter Lists in dictionary_category
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'IHC Breast Cancer Report Intensity', NOW(), 'ihcBcrInte', 'ihc_breast_cancer_report_intensity') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'IHC Breast Cancer Report CerbB2 Pattern', NOW(), 'ihcBcrCerb', 'ihc_breast_cancer_report_cerbb2_pattern') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'IHC Breast Cancer Report Molecular subtype', NOW(), 'ihcBcrMole', 'ihc_breast_cancer_report_molecular_subtype') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::7::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
