-- source: liquibase liquibase/3.5.x.x/003-calendar-management.xml::tat-003-01::tat-module
-- +goose Up
-- +goose StatementBegin
-- Create public_holiday table for TAT Working Time calculations
CREATE SEQUENCE  IF NOT EXISTS clinlims.public_holiday_seq START WITH 1;
CREATE TABLE IF NOT EXISTS clinlims.public_holiday (id INTEGER DEFAULT nextval('public_holiday_seq') NOT NULL, holiday_date date NOT NULL, holiday_name VARCHAR(100) NOT NULL, is_recurring BOOLEAN DEFAULT FALSE, is_active BOOLEAN DEFAULT TRUE, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT public_holiday_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-calendar-management.xml::tat-003-01::tat-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
