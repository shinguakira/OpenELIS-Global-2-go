-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-samples-list-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_SAMPLES join table linking notebooks to SampleItem entities.
CREATE TABLE IF NOT EXISTS notebook_samples_list (notebook_id INTEGER NOT NULL, sample_item_id INTEGER NOT NULL, CONSTRAINT fk_samples_list_sampleitem FOREIGN KEY (sample_item_id) REFERENCES sample_item(id), CONSTRAINT fk_samples_list_notebook FOREIGN KEY (notebook_id) REFERENCES notebook(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_samples_list;
-- +goose StatementEnd
