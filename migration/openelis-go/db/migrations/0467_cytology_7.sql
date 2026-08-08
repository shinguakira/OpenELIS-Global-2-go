-- source: liquibase liquibase/2.8.x.x/cytology.xml::7::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- adds Cytology adequacy in dictionary_category
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Specimen Adequacy - Satisfactory', NOW(), 'AdqSatis', 'cytology_adequacy_satisfactory') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Cytology Specimen Adequacy - Un Satisfactory', NOW(), 'AdqUnSatis', 'cytology_adequacy_unsatisfactory') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/cytology.xml::7::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
