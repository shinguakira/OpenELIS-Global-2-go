-- source: liquibase liquibase/2.6.x.x/eorder_search_enhancements.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- create indices on older tables
CREATE INDEX IF NOT EXISTS e_order_external_id_idx on clinlims.electronic_order (
            external_id
        );

CREATE INDEX IF NOT EXISTS e_order_patient_id_idx on clinlims.electronic_order (
            patient_id
        );

CREATE INDEX IF NOT EXISTS person_lwr_last_name_idx on clinlims.person (
            lower(last_name)
        );

CREATE INDEX IF NOT EXISTS person_lwr_first_name_idx on clinlims.person (
            lower(first_name)
        );

CREATE INDEX IF NOT EXISTS patient_lwr_national_id_idx on clinlims.patient (
           lower(national_id)
        );

CREATE INDEX IF NOT EXISTS patient_identity_identity_data_idx on clinlims.patient_identity (
            lower(identity_data)
        );

CREATE INDEX IF NOT EXISTS patient_identity_patient_id_idx on clinlims.patient_identity (
            patient_id
        );
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/eorder_search_enhancements.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
