-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-tags-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_TAGS table to store tags (as an element collection) associated with each notebook.
CREATE TABLE IF NOT EXISTS notebook_tags (notebook_id INTEGER NOT NULL, tag VARCHAR(255), CONSTRAINT fk_tags_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_tags;
-- +goose StatementEnd
