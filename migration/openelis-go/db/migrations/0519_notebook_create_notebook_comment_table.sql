-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-comment-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_COMMENT table to store comments associated with a notebook.
CREATE TABLE IF NOT EXISTS notebook_comment (id INTEGER NOT NULL, comment_text TEXT, date_created TIMESTAMP WITHOUT TIME ZONE, last_updated TIMESTAMP WITHOUT TIME ZONE, notebook_id INTEGER NOT NULL, system_user_id numeric(10) NOT NULL, CONSTRAINT pk_notebook_comment PRIMARY KEY (id), CONSTRAINT fk_comment_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id), CONSTRAINT fk_comment_system_user FOREIGN KEY (system_user_id) REFERENCES system_user(id));
CREATE SEQUENCE  IF NOT EXISTS notebook_comment_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-comment-table::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
