-- source: liquibase liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-001-label-preset::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Create label_preset table and sequence for configurable label preset config
CREATE SEQUENCE  IF NOT EXISTS clinlims.label_preset_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS clinlims.label_preset (id INTEGER DEFAULT nextval('label_preset_seq') NOT NULL, name VARCHAR(120) NOT NULL, height_mm INTEGER NOT NULL, width_mm INTEGER NOT NULL, barcode_type VARCHAR(20) NOT NULL, prints_per_order BOOLEAN DEFAULT FALSE NOT NULL, prints_per_sample BOOLEAN DEFAULT TRUE NOT NULL, default_per_order INTEGER DEFAULT 0 NOT NULL, max_per_order INTEGER DEFAULT 10 NOT NULL, default_per_sample INTEGER DEFAULT 0 NOT NULL, max_per_sample INTEGER DEFAULT 10 NOT NULL, is_system BOOLEAN DEFAULT FALSE NOT NULL, is_active BOOLEAN DEFAULT TRUE NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT label_preset_pkey PRIMARY KEY (id), CONSTRAINT label_preset_name_uniq UNIQUE (name));
ALTER TABLE clinlims.label_preset
             ADD CONSTRAINT label_preset_height_range  CHECK (height_mm BETWEEN 5 AND 200),
             ADD CONSTRAINT label_preset_width_range   CHECK (width_mm  BETWEEN 5 AND 200),
             ADD CONSTRAINT label_preset_barcode_type  CHECK (barcode_type IN ('CODE_128','QR','DATAMATRIX')),
             ADD CONSTRAINT label_preset_default_nonneg
                 CHECK (default_per_order >= 0 AND default_per_sample >= 0),
             ADD CONSTRAINT label_preset_max_gte_default
                 CHECK (max_per_order >= default_per_order AND max_per_sample >= default_per_sample),
             ADD CONSTRAINT label_preset_scope_required
                 CHECK (prints_per_order OR prints_per_sample);
CREATE INDEX IF NOT EXISTS idx_label_preset_active ON clinlims.label_preset(is_active);
CREATE INDEX IF NOT EXISTS idx_label_preset_system ON clinlims.label_preset(is_system);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/029-label-preset-tables.xml::label-preset-001-label-preset::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
