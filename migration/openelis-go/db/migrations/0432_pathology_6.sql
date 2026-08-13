-- source: liquibase liquibase/2.8.x.x/pathology.xml::6::csteele
-- +goose Up
-- +goose StatementBegin
-- adds in dictionary_type
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Pathology - Techniques', NOW(), 'PathTech', 'pathology_techniques') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Pathology - Pathologist Requests', NOW(), 'PathReq', 'pathologist_requests') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'Pathology - Conclusions', NOW(), 'PathCon', 'pathologist_conclusions') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::6::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
