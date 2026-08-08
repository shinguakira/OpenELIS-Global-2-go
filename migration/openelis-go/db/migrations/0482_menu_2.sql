-- source: liquibase liquibase/2.8.x.x/menu.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
-- update nonconform tab menu option
UPDATE clinlims.menu SET is_active = (SELECT (si.value = 'true') FROM clinlims.site_information si WHERE si.name = 'Non Conformity tab') WHERE element_id IN ('menu_nonconformity', 'menu_non_conforming_report', 'menu_non_conforming_view', 'menu_non_conforming_view');

DELETE FROM clinlims.site_information WHERE name='Non Conformity tab';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
