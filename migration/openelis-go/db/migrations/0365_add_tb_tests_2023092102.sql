-- source: liquibase liquibase/2.7.x.x/add_tb_tests.xml::2023092102::CIV developer Group
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.tb_method_test (id INTEGER DEFAULT nextval('tb_method_test_seq') NOT NULL, test_id VARCHAR(100) NOT NULL, method_id VARCHAR(100) NOT NULL, is_active VARCHAR(2) DEFAULT 'Y', CONSTRAINT tb_method_test_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.tb_method_test;
-- +goose StatementEnd
