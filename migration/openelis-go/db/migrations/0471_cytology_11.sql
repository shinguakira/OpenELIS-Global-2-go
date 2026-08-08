-- source: liquibase liquibase/2.8.x.x/cytology.xml::11::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- adds Cytology Diagnosis Results in dictionary_category
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Epithelial Cell Abnomality - Squamous', NOW(), 'diagEpithS', 'cytology_epithelial_cell_abnomalit_squamous') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Epithelial Cell Abnomality - Glandular', NOW(), 'diagEpithG', 'cytology_epithelial_cell_abnomalit_glandular') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Non-neoplastic cellular variations', NOW(), 'daigNonNeo', 'cytology_non-neoplastic_cellular_variations') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Reactive cellular changes', NOW(), 'daiagReact', 'cytology_reactive_cellular_changes') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Diagnosis Organisms', NOW(), 'diagOrg', 'cytology_diagnosis_organisms') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Diagnosis Other', NOW(), 'diagOther', 'cytology_diagnosis_other') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::11::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
