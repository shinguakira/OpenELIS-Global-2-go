-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating the main NOTEBOOK table to store experiment notebooks.
CREATE TABLE IF NOT EXISTS notebook (id INTEGER NOT NULL, title VARCHAR(255), type numeric(10), objective VARCHAR(255), protocol VARCHAR(255), content TEXT, status VARCHAR(255), is_template BOOLEAN, technician_id numeric(10), creator_id numeric(10), questionnaire_fhir_uuid UUID, last_updated TIMESTAMP WITHOUT TIME ZONE, date_created TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_notebook PRIMARY KEY (id), CONSTRAINT fk_notebook_type_dictionary FOREIGN KEY (type) REFERENCES dictionary(id), CONSTRAINT fk_notebook_technician FOREIGN KEY (technician_id) REFERENCES system_user(id), CONSTRAINT fk_notebook_creator FOREIGN KEY (creator_id) REFERENCES system_user(id));
CREATE SEQUENCE  IF NOT EXISTS notebook_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-table::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
