-- source: liquibase liquibase/2.6.x.x/add_lab_roles_map_table.xml::1::mosesmutesa
-- +goose Up
-- +goose StatementBegin
-- Creating user_lab_unit_roles table
CREATE TABLE IF NOT EXISTS user_lab_unit_roles (system_user_id INTEGER NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT user_lab_unit_roles_pkey PRIMARY KEY (system_user_id));
ALTER TABLE user_lab_unit_roles ADD CONSTRAINT "user_lab_unit_role_systemUser_fk" FOREIGN KEY (system_user_id) REFERENCES system_user (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_lab_roles_map_table.xml::1::mosesmutesa
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
