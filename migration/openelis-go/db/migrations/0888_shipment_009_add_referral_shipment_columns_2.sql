-- source: liquibase liquibase/3.5.x.x/shipment-009-add-referral-shipment-columns.xml::2::pkomena
-- +goose Up
-- +goose StatementBegin
-- Add indexes for shipment tracking queries
CREATE INDEX IF NOT EXISTS idx_referral_assigned_box ON referral(assigned_to_box_id);
CREATE INDEX IF NOT EXISTS idx_referral_lost_status ON referral(lost_status);
CREATE INDEX IF NOT EXISTS idx_referral_priority ON referral(priority);
CREATE INDEX IF NOT EXISTS idx_referral_unassigned ON referral(assigned_to_box_id, lost_status, status, referral_request_date);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/shipment-009-add-referral-shipment-columns.xml::2::pkomena
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
