-- source: liquibase liquibase/3.5.x.x/003-calendar-management.xml::tat-003-03::tat-module
-- +goose Up
-- +goose StatementBegin
-- Add CalendarManagement system module
INSERT INTO clinlims.system_module (id, name, description, has_select_flag, has_add_flag, has_update_flag, has_delete_flag) VALUES (nextval('clinlims.system_module_seq'), 'CalendarManagement', 'Calendar management for public holidays and weekend configuration', 'Y', 'Y', 'Y', 'Y') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/003-calendar-management.xml::tat-003-03::tat-module
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
