-- source: liquibase liquibase/2.2.x.x/email_sms_expansion.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.notification_payload_template DROP CONSTRAINT unique_notification_payload_template_type;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.2.x.x/email_sms_expansion.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
