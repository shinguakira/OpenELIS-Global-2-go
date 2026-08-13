-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-threshold-profile-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create threshold_profile table for reusable temperature/humidity thresholds
CREATE SEQUENCE  IF NOT EXISTS clinlims.threshold_profile_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.threshold_profile (id BIGINT NOT NULL, name VARCHAR(128) NOT NULL, description TEXT, warning_min DECIMAL, warning_max DECIMAL, critical_min DECIMAL, critical_max DECIMAL, min_excursion_minutes INTEGER DEFAULT 5, max_duration_minutes INTEGER, humidity_warning_min DECIMAL, humidity_warning_max DECIMAL, humidity_critical_min DECIMAL, humidity_critical_max DECIMAL, created_by INTEGER NOT NULL, created_at TIMESTAMP WITH TIME ZONE NOT NULL, updated_at TIMESTAMP WITH TIME ZONE, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT threshold_profile_pkey PRIMARY KEY (id), UNIQUE (name));
ALTER TABLE clinlims.threshold_profile ADD CONSTRAINT fk_threshold_profile_created_by FOREIGN KEY (created_by) REFERENCES clinlims.system_user (id) ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-threshold-profile-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
