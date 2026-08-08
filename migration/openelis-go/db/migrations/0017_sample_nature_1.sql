-- source: liquibase liquibase/2.1.x.x/sample_nature.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.dictionary_category (id, description, name, lastupdated, local_abbrev) VALUES (nextval('clinlims.dictionary_category_seq'), 'specimen nature', 'specimen nature', NOW(), 'natureSamp') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.observation_history_type (id, description, type_name, lastupdated) VALUES (nextval('clinlims.observation_history_type_seq'), 'the nature of the sample that was extracted', 'sampleNature', NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/sample_nature.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
