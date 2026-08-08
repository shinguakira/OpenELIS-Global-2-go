-- source: liquibase liquibase/2.2.x.x/email_sms_expansion.xml::3::csteele
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.additional_contacts (contact VARCHAR(255), notification_config_option_id INTEGER, CONSTRAINT fk_additional_contacts_notification_config_option_id FOREIGN KEY (notification_config_option_id) REFERENCES notification_config_option(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clinlims.additional_contacts;
-- +goose StatementEnd
