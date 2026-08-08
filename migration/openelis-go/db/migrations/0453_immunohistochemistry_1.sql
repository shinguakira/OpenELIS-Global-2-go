-- source: liquibase liquibase/2.8.x.x/immunohistochemistry.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create immunohistochemistry table
CREATE TABLE IF NOT EXISTS immunohistochemistry_sample (id INTEGER NOT NULL, technician_id numeric(10), pathologist_id numeric(10), program_id numeric(10), sample_id numeric(10), status VARCHAR(255), questionnaire_response_uuid UUID, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT immunohistochemistry_sample_pkey PRIMARY KEY (id));
ALTER TABLE immunohistochemistry_sample ADD CONSTRAINT immunohistochemistry_sample_program_id_fk FOREIGN KEY (program_id) REFERENCES program (id);
ALTER TABLE immunohistochemistry_sample ADD CONSTRAINT immunohistochemistry_sample_sample_id_fk FOREIGN KEY (sample_id) REFERENCES sample (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/immunohistochemistry.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
