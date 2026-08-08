-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-page-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_PAGE table to store individual pages within a notebook.
CREATE TABLE IF NOT EXISTS notebook_page (id INTEGER NOT NULL, title VARCHAR(255), last_updated TIMESTAMP WITHOUT TIME ZONE, instructions TEXT, content TEXT, page_order INTEGER, completed BOOLEAN DEFAULT FALSE, sample_type_id numeric(10), notebook_id INTEGER NOT NULL, CONSTRAINT pk_notebook_page PRIMARY KEY (id), CONSTRAINT fk_page_sample_type FOREIGN KEY (sample_type_id) REFERENCES type_of_sample(id), CONSTRAINT fk_page_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id));
CREATE SEQUENCE  IF NOT EXISTS notebook_page_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-page-table::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
