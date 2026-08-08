-- source: liquibase liquibase/2.3.x.x/referral_results.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.referral ADD IF NOT EXISTS status VARCHAR(255);
UPDATE clinlims.referral SET status = 'FINISHED' WHERE canceled = false;
UPDATE clinlims.referral SET status = 'CANCELED' WHERE canceled = true;
UPDATE clinlims.referral SET status = 'SENT' WHERE result_recieved_date is NULL AND canceled = false;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/referral_results.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
