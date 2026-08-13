-- source: liquibase liquibase/2.6.x.x/add_lab_roles_map_table.xml::2312131::CIV developer Group
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.lab_unit_roles DROP CONSTRAINT IF EXISTS user_lab_unit_role_map_fk;

ALTER TABLE clinlims.lab_unit_roles ADD CONSTRAINT user_lab_unit_role_map_fk FOREIGN KEY (lab_unit_role_map_id) REFERENCES clinlims.lab_unit_role_map(lab_unit_role_map_id) ON DELETE CASCADE ON UPDATE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/add_lab_roles_map_table.xml::2312131::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
