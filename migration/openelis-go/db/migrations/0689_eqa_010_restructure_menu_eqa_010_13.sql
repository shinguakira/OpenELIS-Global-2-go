-- source: liquibase liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-13::eqa-module
-- +goose Up
-- +goose StatementBegin
-- Add localization entries for restructured EQA menu items
INSERT INTO clinlims.localization (id, description, english, french, lastupdated)
      VALUES
      (nextval('localization_seq'), 'banner.menu.eqa.tests', 'EQA Tests', 'Tests EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tests.tooltip', 'EQA test orders and program enrollment', 'Commandes de tests EQE et inscription aux programmes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tests.orders', 'Orders', 'Commandes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tests.orders.tooltip', 'View and manage EQA test orders', 'Voir et gérer les commandes de tests EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tests.myPrograms', 'My Programs', 'Mes programmes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.tests.myPrograms.tooltip', 'Manage self-enrollment in external EQA programs', 'Gérer l''inscription aux programmes EQE externes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt', 'EQA Management', 'Gestion EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.tooltip', 'Manage EQA programs, participants, and distributions', 'Gérer les programmes EQE, participants et distributions', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.programs', 'Programs', 'Programmes', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.programs.tooltip', 'Create and manage EQA programs', 'Créer et gérer les programmes EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.participants', 'Participants', 'Participants', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.participants.tooltip', 'Manage organization enrollment in EQA programs', 'Gérer l''inscription des organisations aux programmes EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.distributions', 'Distributions', 'Distributions', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.distributions.tooltip', 'Create and manage EQA distributions', 'Créer et gérer les distributions EQE', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.results', 'Results & Analysis', 'Résultats et analyse', CURRENT_TIMESTAMP),
      (nextval('localization_seq'), 'banner.menu.eqa.mgmt.results.tooltip', 'EQA results collection and statistical analysis', 'Collecte de résultats et analyse statistique EQE', CURRENT_TIMESTAMP) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/eqa-010-restructure-menu.xml::eqa-010-13::eqa-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
