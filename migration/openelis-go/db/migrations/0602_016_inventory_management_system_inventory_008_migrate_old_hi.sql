-- source: liquibase liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-008-migrate-old-hiv-kit-data::mherman22
-- +goose Up
-- +goose StatementBegin
-- Migrate old HIV/Syphilis kit data to new schema
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
INSERT INTO clinlims.inventory_item (
                fhir_uuid, name, description, item_type,
                units, is_active, kit_test_type
            )
            SELECT
                uuid_generate_v4(),
                COALESCE(description, 'Unknown Kit'),
                description,
                CASE
                    WHEN UPPER(COALESCE(description, '')) LIKE '%HIV%' THEN 'HIV_KIT'
                    WHEN UPPER(COALESCE(description, '')) LIKE '%SYPHILIS%'
                         OR UPPER(COALESCE(description, '')) LIKE '%SYPHILLIS%' THEN 'SYPHILIS_KIT'
                    ELSE 'HIV_KIT'
                END,
                'kits',
                COALESCE(is_active, 'Y'),
                CASE
                    WHEN UPPER(COALESCE(description, '')) LIKE '%HIV%' THEN 'HIV'
                    WHEN UPPER(COALESCE(description, '')) LIKE '%SYPHILIS%'
                         OR UPPER(COALESCE(description, '')) LIKE '%SYPHILLIS%' THEN 'SYPHILIS'
                    ELSE 'HIV'
                END
            FROM clinlims.inventory_item_backup_20251207 ON CONFLICT DO NOTHING;
INSERT INTO clinlims.inventory_lot (
                fhir_uuid, inventory_item_id, lot_number,
                expiration_date, receipt_date, initial_quantity, current_quantity,
                qc_status, status
            )
            SELECT
                uuid_generate_v4(),
                new_item.id,
                COALESCE(old_loc.lot_number, 'UNKNOWN'),
                old_loc.expiration_date,
                NOW(),
                COALESCE(old_loc.quantity_onhand, 100),
                COALESCE(old_loc.quantity_onhand, 100),
                'PASSED',
                CASE
                    WHEN old_loc.expiration_date IS NOT NULL
                         AND old_loc.expiration_date < NOW() THEN 'EXPIRED'
                    ELSE 'ACTIVE'
                END
            FROM clinlims.inventory_location_backup_20251207 old_loc
            INNER JOIN clinlims.inventory_item_backup_20251207 old_item
                ON old_loc.inv_item_id = old_item.id
            INNER JOIN clinlims.inventory_item new_item
                ON new_item.name = COALESCE(old_item.description, 'Unknown Kit') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/016-inventory-management-system.xml::inventory-008-migrate-old-hiv-kit-data::mherman22
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
