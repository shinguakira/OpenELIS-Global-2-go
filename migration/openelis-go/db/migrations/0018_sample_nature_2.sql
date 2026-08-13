-- source: liquibase liquibase/2.1.x.x/sample_nature.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.dictionary (id, is_active, dict_entry, lastupdated, local_abbrev, dictionary_category_id, display_key, sort_order) VALUES (nextval('clinlims.dictionary_seq'), 'Y', 'Ante mortum', NOW(), 'AnteMorte', (SELECT id FROM dictionary_category WHERE name = 'specimen nature'), 'dictionary.sampleNature.antemortum', '10') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, dict_entry, lastupdated, local_abbrev, dictionary_category_id, display_key, sort_order) VALUES (nextval('clinlims.dictionary_seq'), 'Y', 'Post mortum', NOW(), 'PostMorte', (SELECT id FROM dictionary_category WHERE name = 'specimen nature'), 'dictionary.sampleNature.postmortum', '20') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary (id, is_active, dict_entry, lastupdated, local_abbrev, dictionary_category_id, display_key, sort_order) VALUES (nextval('clinlims.dictionary_seq'), 'Y', 'Environmental', NOW(), 'Environmental', (SELECT id FROM dictionary_category WHERE name = 'specimen nature'), 'dictionary.sampleNature.Environmental', '30') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/sample_nature.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
