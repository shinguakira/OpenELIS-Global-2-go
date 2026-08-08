-- source: liquibase liquibase/2.6.x.x/add_order_priority.xml::4::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update Results Role to have Acces to WorkPlan By Priority Tab
INSERT INTO clinlims.system_module_url (id, url_path, system_module_id) VALUES (nextval('clinlims.system_module_url_seq'), '/WorkPlanByPriority', (SELECT id FROM clinlims.system_module WHERE name = 'Workplan')) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_order_priority.xml::4::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
