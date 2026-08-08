-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-threshold-profile-table::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create freezer_threshold_profile table for time-based threshold assignments
CREATE SEQUENCE  IF NOT EXISTS clinlims.freezer_threshold_profile_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.freezer_threshold_profile (id BIGINT NOT NULL, freezer_id BIGINT NOT NULL, threshold_profile_id BIGINT NOT NULL, effective_start TIMESTAMP WITH TIME ZONE NOT NULL, effective_end TIMESTAMP WITH TIME ZONE, is_default BOOLEAN DEFAULT FALSE, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT freezer_threshold_profile_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.freezer_threshold_profile ADD CONSTRAINT fk_freezer_threshold_freezer FOREIGN KEY (freezer_id) REFERENCES clinlims.freezer (id) ON DELETE CASCADE;
ALTER TABLE clinlims.freezer_threshold_profile ADD CONSTRAINT fk_freezer_threshold_profile FOREIGN KEY (threshold_profile_id) REFERENCES clinlims.threshold_profile (id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-freezer-threshold-profile-table::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
