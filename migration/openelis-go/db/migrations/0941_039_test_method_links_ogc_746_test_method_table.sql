-- source: liquibase liquibase/3.5.x.x/039-test-method-links.xml::OGC-746-test-method-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_method (id VARCHAR(36) NOT NULL, test_id VARCHAR(36) NOT NULL, method_id VARCHAR(36) NOT NULL, is_default BOOLEAN DEFAULT FALSE NOT NULL, effective_date date NOT NULL, is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_method PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.test_method;
-- +goose StatementEnd
