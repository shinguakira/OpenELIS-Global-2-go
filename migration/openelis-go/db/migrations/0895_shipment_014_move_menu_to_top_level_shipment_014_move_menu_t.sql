-- source: liquibase liquibase/3.5.x.x/shipment-014-move-menu-to-top-level.xml::shipment-014-move-menu-to-top-level::pkomena
-- +goose Up
-- +goose StatementBegin
-- Move Sample Shipment menu to top-level, positioned after Patient Management
UPDATE clinlims.menu
            SET parent_id = NULL, presentation_order = 25
            WHERE element_id = 'menu_sample_shipment';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-014-move-menu-to-top-level.xml::shipment-014-move-menu-to-top-level::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
