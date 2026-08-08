-- source: liquibase liquibase/3.5.x.x/003-calendar-management.xml::tat-003-02::tat-module
-- +goose Up
-- +goose StatementBegin
-- Create weekend_config table with seed data (Sat+Sun as default weekends)
CREATE SEQUENCE  IF NOT EXISTS clinlims.weekend_config_seq START WITH 1;
CREATE TABLE IF NOT EXISTS clinlims.weekend_config (id INTEGER DEFAULT nextval('weekend_config_seq') NOT NULL, day_of_week INTEGER NOT NULL, is_weekend BOOLEAN DEFAULT FALSE, sys_user_id INTEGER NOT NULL, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT weekend_config_pkey PRIMARY KEY (id), UNIQUE (day_of_week));
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('0', TRUE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('1', FALSE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('2', FALSE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('3', FALSE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('4', FALSE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('5', FALSE, '1') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.weekend_config (day_of_week, is_weekend, sys_user_id) VALUES ('6', TRUE, '1') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-calendar-management.xml::tat-003-02::tat-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
