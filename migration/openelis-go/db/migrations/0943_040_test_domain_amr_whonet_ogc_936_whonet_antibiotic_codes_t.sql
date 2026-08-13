-- source: liquibase liquibase/3.5.x.x/040-test-domain-amr-whonet.xml::OGC-936-whonet-antibiotic-codes-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.whonet_antibiotic_codes (code VARCHAR(20) NOT NULL, name VARCHAR(255) NOT NULL, antibiotic_class VARCHAR(255), is_active VARCHAR(2) DEFAULT 'Y' NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_whonet_antibiotic_codes PRIMARY KEY (code));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.whonet_antibiotic_codes;
-- +goose StatementEnd
