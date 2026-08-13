-- source: liquibase liquibase/3.3.x.x/030-seed-system-presets.xml::label-preset-seed-001-system-presets::ogc-285
-- +goose Up
-- +goose StatementBegin
-- Seed 5 system label presets from legacy site_information.barcode.* keys per FRS §2.7
WITH si AS (
                SELECT name, value FROM clinlims.site_information
                WHERE name LIKE 'barcode.order.%' OR name LIKE 'barcode.specimen.%'
                   OR name LIKE 'barcode.block.%' OR name LIKE 'barcode.slide.%'
                   OR name LIKE 'barcode.freezer.%'
            ),
            normalized AS (
                SELECT
                    CASE WHEN value ~ '^[0-9]+$' THEN value::INTEGER ELSE NULL END AS num,
                    name
                FROM si
            )
            INSERT INTO clinlims.label_preset
                (id, name, height_mm, width_mm, barcode_type,
                 prints_per_order, prints_per_sample,
                 default_per_order, max_per_order,
                 default_per_sample, max_per_sample,
                 is_system, is_active, last_updated)
            SELECT
                nextval('clinlims.label_preset_seq'),
                preset_name,
                COALESCE((SELECT num FROM normalized WHERE name = 'barcode.' || legacy_key || '.height'), 25)  AS height_mm,
                COALESCE((SELECT num FROM normalized WHERE name = 'barcode.' || legacy_key || '.width'),  76)  AS width_mm,
                'CODE_128',
                (legacy_key = 'order')      AS prints_per_order,
                (legacy_key <> 'order') AS prints_per_sample,
                CASE WHEN legacy_key = 'order'      THEN COALESCE((SELECT num FROM normalized WHERE name = 'barcode.order.default'), 1) ELSE 0 END,
                CASE WHEN legacy_key = 'order'      THEN COALESCE((SELECT num FROM normalized WHERE name = 'barcode.order.max'),    10) ELSE 10 END,
                CASE WHEN legacy_key <> 'order' THEN COALESCE((SELECT num FROM normalized WHERE name = 'barcode.' || legacy_key || '.default'), 1) ELSE 0 END,
                CASE WHEN legacy_key <> 'order' THEN COALESCE((SELECT num FROM normalized WHERE name = 'barcode.' || legacy_key || '.max'),    10) ELSE 10 END,
                true,
                true,
                CURRENT_TIMESTAMP
            FROM (VALUES
                ('order',    'Order Label'),
                ('specimen', 'Specimen Label'),
                ('block',    'Block Label'),
                ('slide',    'Slide Label'),
                ('freezer',  'Freezer Label')
            ) AS seed(legacy_key, preset_name);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/030-seed-system-presets.xml::label-preset-seed-001-system-presets::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
