-- source: liquibase liquibase/3.3.x.x/032-label-preset-is-universal.xml::label-preset-is-universal-001::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Add label_preset.is_universal (FR-014a) — per-sample presets that always emit a column
ALTER TABLE clinlims.label_preset ADD IF NOT EXISTS is_universal BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE clinlims.label_preset
             ADD CONSTRAINT label_preset_universal_per_sample
                 CHECK (NOT is_universal OR prints_per_sample);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/032-label-preset-is-universal.xml::label-preset-is-universal-001::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
