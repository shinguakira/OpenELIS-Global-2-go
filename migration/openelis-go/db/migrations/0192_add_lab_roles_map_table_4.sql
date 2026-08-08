-- source: liquibase liquibase/2.6.x.x/add_lab_roles_map_table.xml::4::mosesmutesa
-- +goose Up
-- +goose StatementBegin
-- Creating lab_unit_roles table
CREATE TABLE IF NOT EXISTS lab_unit_roles (system_user_id INTEGER NOT NULL, lab_unit_role_map_id INTEGER NOT NULL);
ALTER TABLE lab_unit_roles ADD CONSTRAINT user_lab_role_id__fk FOREIGN KEY (system_user_id) REFERENCES user_lab_unit_roles (system_user_id);
ALTER TABLE lab_unit_roles ADD CONSTRAINT user_lab_unit_role_map_fk FOREIGN KEY (lab_unit_role_map_id) REFERENCES lab_unit_role_map (lab_unit_role_map_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_lab_roles_map_table.xml::4::mosesmutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
