-- source: liquibase liquibase/2.6.x.x/add_lab_roles_map_table.xml::3::mosesmutesa
-- +goose Up
-- +goose StatementBegin
-- Creating lab_roles table
CREATE TABLE IF NOT EXISTS lab_roles (lab_unit_role_map_id INTEGER NOT NULL, role VARCHAR(64));
ALTER TABLE lab_roles ADD CONSTRAINT lab_unit_role_map__fk FOREIGN KEY (lab_unit_role_map_id) REFERENCES lab_unit_role_map (lab_unit_role_map_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_lab_roles_map_table.xml::3::mosesmutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
