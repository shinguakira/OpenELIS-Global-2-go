-- source: liquibase liquibase/2.6.x.x/add_order_priority.xml::5::moses_mutesa
-- +goose Up
-- +goose StatementBegin
-- update correct WorkPlanByPriority action url
UPDATE clinlims.menu SET action_url = '/WorkPlanByPriority?type=priority' WHERE element_id='menu_workplan_priority';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_order_priority.xml::5::moses_mutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
