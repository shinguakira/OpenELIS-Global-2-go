-- source: liquibase liquibase/2.8.x.x/menu.xml::10::csteele
-- +goose Up
-- +goose StatementBegin
-- update order of top level menu
UPDATE clinlims.menu SET presentation_order = '0' WHERE element_id='menu_home';

UPDATE clinlims.menu SET presentation_order = '10' WHERE element_id='menu_sample';

UPDATE clinlims.menu SET presentation_order = '20' WHERE element_id='menu_patient';

UPDATE clinlims.menu SET presentation_order = '30' WHERE element_id='menu_nonconformity';

UPDATE clinlims.menu SET presentation_order = '40' WHERE element_id='menu_workplan';

UPDATE clinlims.menu SET presentation_order = '50' WHERE element_id='menu_pathology';

UPDATE clinlims.menu SET presentation_order = '60' WHERE element_id='menu_immunochem';

UPDATE clinlims.menu SET presentation_order = '70' WHERE element_id='menu_cytology';

UPDATE clinlims.menu SET presentation_order = '80' WHERE element_id='menu_results';

UPDATE clinlims.menu SET presentation_order = '90' WHERE element_id='menu_resultvalidation';

UPDATE clinlims.menu SET presentation_order = '100' WHERE element_id='menu_reports';

UPDATE clinlims.menu SET presentation_order = '110' WHERE element_id='menu_administration';

UPDATE clinlims.menu SET presentation_order = '120' WHERE element_id='menu_billing';

UPDATE clinlims.menu SET presentation_order = '130' WHERE element_id='menu_help';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/menu.xml::10::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
