-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-samples-table-full::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_SAMPLES table as entity table (not join table) to store notebook sample relationships with questionnaire responses.
CREATE TABLE IF NOT EXISTS notebook_samples (id INTEGER NOT NULL, notebook_id INTEGER NOT NULL, sample_item_id INTEGER NOT NULL, questionnaire_response_uuid UUID, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_notebook_samples PRIMARY KEY (id), CONSTRAINT fk_samples_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id), CONSTRAINT fk_samples_sampleitem FOREIGN KEY (sample_item_id) REFERENCES sample_item(id));
CREATE SEQUENCE  IF NOT EXISTS notebook_samples_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-samples-table-full::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
