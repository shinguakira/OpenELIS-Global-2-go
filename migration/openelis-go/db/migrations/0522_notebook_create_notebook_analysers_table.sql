-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-analysers-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_ANALYSERS join table linking notebooks to Analyzer entities.
CREATE TABLE IF NOT EXISTS notebook_analysers (notebook_id INTEGER NOT NULL, analyser_id INTEGER NOT NULL, CONSTRAINT fk_analysers_analyser FOREIGN KEY (analyser_id) REFERENCES analyzer(id), CONSTRAINT fk_analysers_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_analysers;
-- +goose StatementEnd
