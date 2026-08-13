-- source: liquibase liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-002-label-preset-field::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Create label_preset_field table and sequence for per-preset content field ordering
CREATE SEQUENCE  IF NOT EXISTS clinlims.label_preset_field_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.label_preset_field (id INTEGER DEFAULT nextval('label_preset_field_seq') NOT NULL, preset_id INTEGER NOT NULL, field_key VARCHAR(60) NOT NULL, source_type VARCHAR(20) DEFAULT 'SYSTEM' NOT NULL, is_required BOOLEAN DEFAULT FALSE NOT NULL, display_order INTEGER NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT label_preset_field_pkey PRIMARY KEY (id), CONSTRAINT fk_label_preset_field_preset FOREIGN KEY (preset_id) REFERENCES label_preset(id) ON DELETE CASCADE);
ALTER TABLE clinlims.label_preset_field ADD CONSTRAINT label_preset_field_order_uniq UNIQUE (preset_id, display_order);
ALTER TABLE clinlims.label_preset_field ADD CONSTRAINT label_preset_field_key_uniq UNIQUE (preset_id, field_key);
ALTER TABLE clinlims.label_preset_field
             ADD CONSTRAINT label_preset_field_source_type CHECK (source_type = 'SYSTEM');
CREATE INDEX IF NOT EXISTS idx_label_preset_field_preset ON clinlims.label_preset_field(preset_id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-002-label-preset-field::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
