-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-entries-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_ENTRIES join table linking notebooks to their entries.
CREATE TABLE IF NOT EXISTS notebook_entries (notebook_id INTEGER NOT NULL, entry_id INTEGER NOT NULL, CONSTRAINT fk_entries_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id), CONSTRAINT fk_entries_entry FOREIGN KEY (entry_id) REFERENCES notebook(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_entries;
-- +goose StatementEnd
