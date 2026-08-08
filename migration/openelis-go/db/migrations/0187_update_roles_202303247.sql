-- source: liquibase liquibase/2.6.x.x/update_roles.xml::202303247::CIV Developer Group
-- +goose Up
-- +goose StatementBegin
-- update roles for RetroCI context
insert into clinlims.system_role
            values(nextval('clinlims.system_role_seq'),'Reception','Sample entry
            and patient management.',FALSE,(select id from clinlims.system_role
            where name = 'Lab Unit Roles'),'role.intake',TRUE,FALSE) ON CONFLICT DO NOTHING;
update
            clinlims.system_role set active = TRUE where name in ('Global
            Roles','Lab Unit Roles');
update clinlims.system_role set
            grouping_parent = (select id from
            clinlims.system_role where name =
            'Global Roles') where name = 'Audit
            Trail';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.6.x.x/update_roles.xml::202303247::CIV Developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
