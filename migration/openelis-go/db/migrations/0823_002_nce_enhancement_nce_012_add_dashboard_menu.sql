-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-012-add-dashboard-menu::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Add NCE Dashboard menu item under Non-Conformity
INSERT INTO clinlims.menu (id, parent_id, presentation_order, element_id, action_url, display_key, is_active)
            SELECT
                COALESCE(MAX(id), 0) + 1,
                (SELECT id FROM clinlims.menu WHERE element_id = 'menu_nonconformity'),
                0,
                'menu_nce_dashboard',
                '/NceDashboard',
                'banner.menu.nonconformity.dashboard',
                true
            FROM clinlims.menu ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-012-add-dashboard-menu::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
