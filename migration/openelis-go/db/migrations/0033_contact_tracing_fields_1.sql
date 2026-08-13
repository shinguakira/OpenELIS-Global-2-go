-- source: liquibase liquibase/2.3.x.x/contact_tracing_fields.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.sample_additional_fields (sample_id numeric(10) NOT NULL, field_name VARCHAR(255) NOT NULL, field_value VARCHAR(255), last_updated date, CONSTRAINT pk_sample_additional_fields PRIMARY KEY (sample_id, field_name), CONSTRAINT fk_sample_additional_fields_sample FOREIGN KEY (sample_id) REFERENCES sample);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.sample_additional_fields;
-- +goose StatementEnd
