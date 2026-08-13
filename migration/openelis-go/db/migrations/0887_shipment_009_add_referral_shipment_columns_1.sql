-- source: liquibase liquibase/3.5.x.x/shipment-009-add-referral-shipment-columns.xml::1::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add shipment tracking columns to referral table for unassigned samples workflow
ALTER TABLE referral ADD IF NOT EXISTS assigned_to_box_id INTEGER;
ALTER TABLE referral ADD CONSTRAINT fk_referral_assigned_box FOREIGN KEY (assigned_to_box_id) REFERENCES shipping_box (id);
ALTER TABLE referral ADD IF NOT EXISTS lost_status BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE referral ADD IF NOT EXISTS lost_date TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE referral ADD IF NOT EXISTS lost_reason TEXT;
ALTER TABLE referral ADD IF NOT EXISTS priority VARCHAR(20);
ALTER TABLE referral ADD IF NOT EXISTS cancel_date TIMESTAMP WITHOUT TIME ZONE;
ALTER TABLE referral ADD IF NOT EXISTS cancel_reason TEXT;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-009-add-referral-shipment-columns.xml::1::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
