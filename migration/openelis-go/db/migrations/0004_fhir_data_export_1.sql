-- source: liquibase liquibase/2.0.x.x/fhir_data_export.xml::1::caleb
-- +goose Up
-- +goose StatementBegin
-- add functionality for fhir data export
CREATE TABLE IF NOT EXISTS clinlims.data_export_task (id BIGINT NOT NULL, data_request_attempt_timeout INTEGER, max_data_export_interval INTEGER, endpoint VARCHAR(255), CONSTRAINT data_export_task_pkey PRIMARY KEY (id), UNIQUE (endpoint));
CREATE TABLE IF NOT EXISTS clinlims.data_export_fhir_resources (data_export_task_id BIGINT, fhir_resources VARCHAR(255), CONSTRAINT fk_data_export_fhir_resources_data_export_task FOREIGN KEY (data_export_task_id) REFERENCES data_export_task(id));
CREATE TABLE IF NOT EXISTS clinlims.data_export_headers (data_export_task_id BIGINT, key VARCHAR(255), value VARCHAR(255), CONSTRAINT fk_data_export_headers_data_export_task FOREIGN KEY (data_export_task_id) REFERENCES data_export_task(id));
CREATE TABLE IF NOT EXISTS clinlims.data_export_attempt (id BIGINT NOT NULL, start_time TIMESTAMP WITHOUT TIME ZONE, end_time TIMESTAMP WITHOUT TIME ZONE, data_export_task_id BIGINT, data_export_status VARCHAR(255), CONSTRAINT data_export_attempt_pkey PRIMARY KEY (id), CONSTRAINT fk_data_export_attempt_data_export_task FOREIGN KEY (data_export_task_id) REFERENCES data_export_task(id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/fhir_data_export.xml::1::caleb
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
