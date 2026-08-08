-- source: liquibase liquibase/2.1.x.x/client_notifications.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.client_results_view (id INTEGER NOT NULL, password VARCHAR(255), result_id numeric(10), last_updated date, CONSTRAINT client_results_view_pkey PRIMARY KEY (id), CONSTRAINT fk_client_results_view_result FOREIGN KEY (result_id) REFERENCES result(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.client_results_view_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/client_notifications.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
