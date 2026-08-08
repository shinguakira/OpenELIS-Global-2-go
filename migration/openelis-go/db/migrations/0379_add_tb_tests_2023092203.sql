-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092203::CIV developer Group
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.tb_method_panel (id INTEGER DEFAULT nextval('tb_method_panel_seq') NOT NULL, panel_id numeric NOT NULL, method_id numeric NOT NULL, is_active VARCHAR(2) DEFAULT 'Y', CONSTRAINT tb_method_panel_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.tb_method_panel;
-- +goose StatementEnd
