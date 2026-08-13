-- source: liquibase liquibase/2.1.x.x/client_notifications.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.notification_payload_template (id INTEGER NOT NULL, message_template VARCHAR(10000), subject_template VARCHAR(255), type VARCHAR(255) NOT NULL, last_updated date, CONSTRAINT notification_payload_template_pkey PRIMARY KEY (id), CONSTRAINT unique_notification_payload_template_type UNIQUE (type));
CREATE SEQUENCE  IF NOT EXISTS clinlims.notification_payload_template_seq START WITH 1 INCREMENT BY 1;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/client_notifications.xml::3::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
