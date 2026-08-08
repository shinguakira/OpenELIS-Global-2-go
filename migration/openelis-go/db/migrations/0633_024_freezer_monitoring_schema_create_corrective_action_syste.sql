-- source: liquibase liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-corrective-action-system::mherman22
-- +goose Up
-- +goose StatementBegin
-- Create device-centric corrective_action table for tracking maintenance and repair actions
--             on cold storage devices (freezers). Actions are linked directly to devices, not alerts.
CREATE SEQUENCE  IF NOT EXISTS clinlims.corrective_action_seq START WITH 1 INCREMENT BY 1 CACHE 1;
CREATE TABLE IF NOT EXISTS clinlims.corrective_action (id BIGINT DEFAULT nextval('corrective_action_seq') NOT NULL, freezer_id BIGINT NOT NULL, action_type VARCHAR(50) NOT NULL, description TEXT NOT NULL, status VARCHAR(20) DEFAULT 'PENDING' NOT NULL, created_by INTEGER NOT NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL, updated_by INTEGER, updated_at TIMESTAMP WITH TIME ZONE, is_edited BOOLEAN DEFAULT FALSE NOT NULL, completed_at TIMESTAMP WITH TIME ZONE, completion_notes TEXT, retracted_at TIMESTAMP WITH TIME ZONE, retraction_reason TEXT, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(), CONSTRAINT corrective_action_pkey PRIMARY KEY (id));
ALTER TABLE clinlims.corrective_action ADD CONSTRAINT fk_corrective_action_freezer FOREIGN KEY (freezer_id) REFERENCES clinlims.freezer (id) ON DELETE CASCADE;
ALTER TABLE clinlims.corrective_action ADD CONSTRAINT fk_corrective_action_created_by FOREIGN KEY (created_by) REFERENCES clinlims.system_user (id) ON DELETE RESTRICT;
ALTER TABLE clinlims.corrective_action ADD CONSTRAINT fk_corrective_action_updated_by FOREIGN KEY (updated_by) REFERENCES clinlims.system_user (id) ON DELETE RESTRICT;
ALTER TABLE clinlims.corrective_action ADD CONSTRAINT chk_corrective_action_type
            CHECK (action_type IN (
            'TEMPERATURE_ADJUSTMENT',
            'EQUIPMENT_REPAIR',
            'SAMPLE_RELOCATION',
            'CALIBRATION',
            'ITEM_REORDER',
            'MAINTENANCE',
            'OTHER'
            ));
ALTER TABLE clinlims.corrective_action ADD CONSTRAINT chk_corrective_action_status
            CHECK (status IN (
            'PENDING',
            'IN_PROGRESS',
            'COMPLETED',
            'CANCELLED',
            'RETRACTED'
            ));
CREATE INDEX IF NOT EXISTS idx_corrective_action_freezer ON clinlims.corrective_action(freezer_id);
CREATE INDEX IF NOT EXISTS idx_corrective_action_status ON clinlims.corrective_action(status);
CREATE INDEX IF NOT EXISTS idx_corrective_action_created_at ON clinlims.corrective_action(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_corrective_action_edited ON clinlims.corrective_action(is_edited);
CREATE INDEX IF NOT EXISTS idx_corrective_action_freezer_created ON clinlims.corrective_action(freezer_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/024-freezer-monitoring-schema.xml::create-corrective-action-system::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
