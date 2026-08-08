-- source: liquibase liquibase/3.3.x.x/025-register-freezer-audit-trail.xml::register-freezer-audit-trail::mherman22
-- +goose Up
-- +goose StatementBegin
-- Register freezer-related tables in reference_tables to enable automatic audit trailing
--             via OpenELIS history table. This allows tracking configuration changes, threshold updates,
--             and other freezer modifications for regulatory compliance.
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded) VALUES (nextval('clinlims.reference_tables_seq'), 'FREEZER', 'Y', 'N') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded) VALUES (nextval('clinlims.reference_tables_seq'), 'CORRECTIVE_ACTION', 'Y', 'N') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.reference_tables (id, name, keep_history, is_hl7_encoded) VALUES (nextval('clinlims.reference_tables_seq'), 'ALERT', 'Y', 'N') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/025-register-freezer-audit-trail.xml::register-freezer-audit-trail::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
