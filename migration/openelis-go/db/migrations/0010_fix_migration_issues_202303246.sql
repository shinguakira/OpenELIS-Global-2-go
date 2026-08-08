-- source: liquibase liquibase/2.1.x.x/fix_migration_issues.xml::202303246::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
-- Update  localisation english label
alter table clinlims.system_role drop constraint
            role_parent_role_fk,add constraint role_parent_role_fk foreign key
            (grouping_parent) references clinlims.system_role (id) MATCH SIMPLE
            ON UPDATE NO ACTION ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.1.x.x/fix_migration_issues.xml::202303246::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
