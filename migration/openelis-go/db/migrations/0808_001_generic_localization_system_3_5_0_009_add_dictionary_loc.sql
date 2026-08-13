-- source: liquibase liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-009-add-dictionary-localization::reagan-meant
-- +goose Up
-- +goose StatementBegin
DO $$
            DECLARE
                dict_record RECORD;
                new_loc_id NUMERIC;
            BEGIN
                FOR dict_record IN
                    SELECT id, dict_entry
                    FROM clinlims.dictionary
                    WHERE name_localization_id IS NULL AND dict_entry IS NOT NULL
                LOOP
                    -- Create localization record (include legacy english/french columns for backward compatibility)
                    SELECT nextval('localization_seq') INTO new_loc_id;

                    INSERT INTO clinlims.localization (id, description, english, french, lastupdated)
                    VALUES (new_loc_id, 'dictionary name', dict_record.dict_entry, dict_record.dict_entry, now());

                    -- Create English translation only (other languages to be added via translation UI)
                    INSERT INTO clinlims.localization_value (id, localization_id, locale, value, last_updated)
                    VALUES (nextval('localization_value_seq'), new_loc_id, 'en', dict_record.dict_entry, now());

                    -- Update dictionary with localization reference
                    UPDATE clinlims.dictionary
                    SET name_localization_id = new_loc_id
                    WHERE id = dict_record.id;
                END LOOP;
            END $$;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/001-generic-localization-system.xml::3.5.0-009-add-dictionary-localization::reagan-meant
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
