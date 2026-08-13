-- source: liquibase liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-05::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Add localization entries for EQA menu items
INSERT INTO clinlims.localization (id, description, english, french, lastupdated)
      VALUES
      (nextval('localization_seq'), 'banner.menu.eqa', 'EQA', 'EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tooltip', 'External Quality Assurance', 'Assurance qualité externe', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.alerts', 'Alerts', 'Alertes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.alerts.tooltip', 'Alerts and quality assurance dashboard', 'Tableau de bord des alertes et assurance qualité', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.management', 'EQA Programs', 'Programmes EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.management.tooltip', 'Create and manage EQA programs', 'Créer et gérer les programmes EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.distribution', 'EQA Distributions', 'Distributions EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.distribution.tooltip', 'Create and manage EQA distributions', 'Créer et gérer les distributions EQE', CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-007-add-eqa-menu-items.xml::eqa-007-05::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
