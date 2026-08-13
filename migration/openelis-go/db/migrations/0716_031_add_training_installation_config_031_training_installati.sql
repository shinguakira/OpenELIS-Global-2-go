-- source: liquibase liquibase/3.3.x.x/031-add-training-installation-config.xml::031-training-installation-1::oe
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.site_information (id, name, description, value, value_type, domain_id, lastupdated) VALUES (nextval('clinlims.site_information_seq'), 'TrainingInstallation', 'If true, allows deletion of all patient data for training purposes. Enable only on training/demo instances.', 'false', 'text', (SELECT id FROM clinlims.site_information_domain WHERE name = 'siteIdentity'), NOW()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/031-add-training-installation-config.xml::031-training-installation-1::oe
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
