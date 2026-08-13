-- source: liquibase liquibase/3.2.x.x/notebook.xml::create-notebook-menu-option::mozzymutesa
-- +goose Up
-- +goose StatementBegin
-- insert NoteBook tab menu option
INSERT INTO clinlims.menu (id, presentation_order, element_id, display_key, tool_tip_key, new_window, is_active, hide_in_old_ui, action_url) VALUES (nextval('clinlims.menu_seq'), '125', 'menu_notebook', 'sidenav.label.notebook', 'sidenav.label.notebook.tooltip', FALSE, TRUE, TRUE, '/NotebookDashboard') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.2.x.x/notebook.xml::create-notebook-menu-option::mozzymutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
