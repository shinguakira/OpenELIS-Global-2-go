-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-page-panels-table::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- Creating NOTEBOOK_PAGE_PANELS table to store panels associated with each notebook page.
CREATE TABLE IF NOT EXISTS notebook_page_panels (notebook_page_id INTEGER NOT NULL, panel numeric(10) NOT NULL, CONSTRAINT fk_panel_id FOREIGN KEY (panel) REFERENCES panel(id), CONSTRAINT fk_panels_notebook_page FOREIGN KEY (notebook_page_id) REFERENCES notebook_page(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notebook_page_panels;
-- +goose StatementEnd
