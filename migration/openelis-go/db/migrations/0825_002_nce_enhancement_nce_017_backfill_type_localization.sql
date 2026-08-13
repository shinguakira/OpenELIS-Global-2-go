-- source: liquibase liquibase/3.5.x.x/002-nce-enhancement.xml::nce-017-backfill-type-localization::reagan-meant
-- +goose Up
-- +goose StatementBegin
-- Create localization records for existing nce_type rows without localization
DO $$
            DECLARE
                type_record RECORD;
                new_loc_id INTEGER;
            BEGIN
                FOR type_record IN
                    SELECT id, name FROM clinlims.nce_type WHERE name_localization_id IS NULL AND name IS NOT NULL
                LOOP
                    -- Insert into localization table
                    INSERT INTO clinlims.localization (id, description, lastupdated)
                    VALUES (nextval('clinlims.localization_seq'), 'NCE Type: ' || type_record.name, NOW())
                    RETURNING id INTO new_loc_id;

                    -- Insert English value
                    INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
                    VALUES (nextval('clinlims.localization_value_seq'), new_loc_id, 'en', type_record.name, NOW());

                    -- Update nce_type with localization reference
                    UPDATE clinlims.nce_type SET name_localization_id = new_loc_id WHERE id = type_record.id;
                END LOOP;
            END $$;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/002-nce-enhancement.xml::nce-017-backfill-type-localization::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
