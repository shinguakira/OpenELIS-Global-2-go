-- source: liquibase liquibase/2.7.x.x/activate_result_modifer_role.xml::23101801::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.system_role SET active = TRUE, editable = FALSE, grouping_parent = (SELECT id FROM system_role WHERE name = 'Lab Unit Roles') WHERE id=(SELECT id FROM system_role WHERE name = 'Results modifier');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/activate_result_modifer_role.xml::23101801::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
