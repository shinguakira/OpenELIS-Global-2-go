-- source: liquibase liquibase/3.3.x.x/031-create-sample-qa-checklist-table.xml::create-qa-checklist-dictionary-category::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Create dictionary category for QA Checklist items.
--             Dictionary entries should be added via CSV configuration file at:
--             volume/configuration/backend/dictionaries/qa-checklist-items.csv
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'QA Checklist items for order verification in Step 4', NOW(), 'QACheck', 'QAChecklistItem') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/031-create-sample-qa-checklist-table.xml::create-qa-checklist-dictionary-category::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
