-- source: liquibase liquibase/3.3.x.x/031-seed-system-preset-fields.xml::label-preset-seed-002-system-preset-fields::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Seed LAB_NUMBER (required, position 1) for each system preset
INSERT INTO clinlims.label_preset_field (id, preset_id, field_key, source_type, is_required, display_order, last_updated)
            SELECT nextval('clinlims.label_preset_field_seq'),
                   id, 'LAB_NUMBER', 'SYSTEM', true, 1, CURRENT_TIMESTAMP
            FROM clinlims.label_preset
            WHERE is_system = true ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/031-seed-system-preset-fields.xml::label-preset-seed-002-system-preset-fields::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
