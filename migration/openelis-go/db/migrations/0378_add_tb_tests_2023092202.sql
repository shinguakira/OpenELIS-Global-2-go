-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092202::CIV developer Group
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.tb_method_panel_seq START WITH 1 INCREMENT BY 1 CACHE 200;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SEQUENCE IF EXISTS clinlims.tb_method_panel_seq;
-- +goose StatementEnd
