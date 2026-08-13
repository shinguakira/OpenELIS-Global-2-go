-- source: liquibase liquibase/3.5.x.x/shipment-006-create-foreign-keys.xml::shipment-006-create-foreign-keys::pkomena
-- +goose Up
-- +goose StatementBegin
ALTER TABLE shipping_box ADD CONSTRAINT fk_shipping_box_destination FOREIGN KEY (destination_facility_id) REFERENCES organization (id) ON UPDATE CASCADE ON DELETE RESTRICT;

ALTER TABLE shipping_box ADD CONSTRAINT fk_shipping_box_created_by FOREIGN KEY (created_by) REFERENCES system_user (id) ON UPDATE CASCADE ON DELETE SET NULL;

ALTER TABLE shipment ADD CONSTRAINT fk_shipment_box FOREIGN KEY (shipping_box_id) REFERENCES shipping_box (id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE box_sample ADD CONSTRAINT fk_box_sample_box FOREIGN KEY (shipping_box_id) REFERENCES shipping_box (id) ON UPDATE CASCADE ON DELETE CASCADE;

ALTER TABLE box_sample ADD CONSTRAINT fk_box_sample_sample FOREIGN KEY (sample_id) REFERENCES sample (id) ON UPDATE CASCADE ON DELETE RESTRICT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-006-create-foreign-keys.xml::shipment-006-create-foreign-keys::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
