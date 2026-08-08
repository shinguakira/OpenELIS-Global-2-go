-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-page-tests-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_PAGE_TESTS table to store tests associated with each notebook page.
CREATE TABLE IF NOT EXISTS notebook_page_tests (notebook_page_id INTEGER NOT NULL, test INTEGER, CONSTRAINT fk_tests_notebook_page FOREIGN KEY (notebook_page_id) REFERENCES notebook_page(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_page_tests;
-- +goose StatementEnd
