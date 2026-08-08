-- source: liquibase liquibase/3.3.x.x/020-storage-box-use-code-not-short-code.xml::storage-020-storage-box-use-code-not-short-code::openelisglobal-ai-agent
-- +goose Up
-- +goose StatementBegin
ALTER TABLE storage_box ADD IF NOT EXISTS code VARCHAR(10);
-- 1) Seed code from short_code (preferred) or generated from label
            UPDATE clinlims.storage_box
            SET code = COALESCE(
                NULLIF(TRIM(short_code), ''),
                LEFT(UPPER(REGEXP_REPLACE(label, '[^A-Z0-9_-]', '', 'g')), 10)
            )
            WHERE code IS NULL;
-- 2) Ensure <= 10 chars
            UPDATE clinlims.storage_box
            SET code = LEFT(code, 10)
            WHERE code IS NOT NULL AND LENGTH(code) > 10;
-- 3) Resolve duplicates within the same parent_rack_id by adding numeric suffix
            WITH ranked AS (
                SELECT
                    id,
                    parent_rack_id,
                    code AS base_code,
                    ROW_NUMBER() OVER (
                        PARTITION BY parent_rack_id, code
                        ORDER BY id
                    ) AS rn
                FROM clinlims.storage_box
                WHERE code IS NOT NULL
            )
            UPDATE clinlims.storage_box b
            SET code = CASE
                WHEN r.rn = 1 THEN r.base_code
                ELSE
                    (
                        LEFT(
                            r.base_code,
                            GREATEST(
                                1,
                                10 - (1 + LENGTH((r.rn - 1)::text))
                            )
                        ) || '-' || (r.rn - 1)::text
                    )
            END
            FROM ranked r
            WHERE b.id = r.id;
ALTER TABLE storage_box ALTER COLUMN  code SET NOT NULL;
ALTER TABLE storage_box ADD CONSTRAINT uk_box_code_in_rack UNIQUE (parent_rack_id, code);
ALTER TABLE storage_box DROP COLUMN IF EXISTS short_code;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/020-storage-box-use-code-not-short-code.xml::storage-020-storage-box-use-code-not-short-code::openelisglobal-ai-agent
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
