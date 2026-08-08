-- source: liquibase liquibase/3.2.x.x/notebook.xml::add-notebook-expt-type-category::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- adds notebook_experiment_typee dictionary Category in dictionary_category table
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'NoteBook Experiment Type', NOW(), 'NoteExpTyp', 'notebook_experiment_type') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::add-notebook-expt-type-category::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
