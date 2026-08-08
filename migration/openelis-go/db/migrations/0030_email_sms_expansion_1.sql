-- source: liquibase liquibase/2.2.x.x/email_sms_expansion.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
CREATE SEQUENCE  IF NOT EXISTS clinlims.notification_config_option_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.notification_config_option (id INTEGER NOT NULL, notification_nature VARCHAR(255), notification_person_type VARCHAR(255), notification_method VARCHAR(255), payload_template_id INTEGER, active BOOLEAN, last_updated date, CONSTRAINT notification_config_option_pkey PRIMARY KEY (id), CONSTRAINT fk_notification_config_option_payload_template FOREIGN KEY (payload_template_id) REFERENCES notification_payload_template(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.test_notification_config_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.test_notification_config (id INTEGER NOT NULL, test_id numeric(10), default_payload_template_id INTEGER, last_updated date, CONSTRAINT test_notification_config_pkey PRIMARY KEY (id), CONSTRAINT fk_test_notification_config_test FOREIGN KEY (test_id) REFERENCES test(id), CONSTRAINT fk_test_notification_config_default_payload_template FOREIGN KEY (default_payload_template_id) REFERENCES notification_payload_template(id), UNIQUE (test_id));
CREATE TABLE IF NOT EXISTS clinlims.test_notification_config_config_option (test_notification_config_id INTEGER NOT NULL, notification_config_option_id INTEGER NOT NULL, CONSTRAINT pk_test_notification_config_config_option PRIMARY KEY (test_notification_config_id, notification_config_option_id), CONSTRAINT fk_test_notification_config_config_option_notification_config FOREIGN KEY (test_notification_config_id) REFERENCES test_notification_config(id), CONSTRAINT fk_test_notification_config_config_option_notification_config_option FOREIGN KEY (notification_config_option_id) REFERENCES notification_config_option(id));
CREATE SEQUENCE  IF NOT EXISTS clinlims.analysis_notification_config_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.analysis_notification_config (id INTEGER NOT NULL, analysis_id numeric(10), default_payload_template_id INTEGER, last_updated date, CONSTRAINT analysis_notification_config_pkey PRIMARY KEY (id), CONSTRAINT fk_analysis_notification_config_default_payload_template FOREIGN KEY (default_payload_template_id) REFERENCES notification_payload_template(id), CONSTRAINT fk_analysis_notification_config_test FOREIGN KEY (analysis_id) REFERENCES analysis(id), UNIQUE (analysis_id));
CREATE TABLE IF NOT EXISTS clinlims.analysis_notification_config_config_option (analysis_notification_config_id INTEGER NOT NULL, notification_config_option_id INTEGER NOT NULL, CONSTRAINT pk_analysis_notification_config_config_option PRIMARY KEY (analysis_notification_config_id, notification_config_option_id), CONSTRAINT fk_analysis_notification_notification_config_option FOREIGN KEY (notification_config_option_id) REFERENCES notification_config_option(id), CONSTRAINT fk_analysis_notification_notification_config FOREIGN KEY (analysis_notification_config_id) REFERENCES analysis_notification_config(id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.2.x.x/email_sms_expansion.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
