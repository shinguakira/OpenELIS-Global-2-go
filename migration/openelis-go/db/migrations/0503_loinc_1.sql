-- source: liquibase liquibase/2.8.x.x/loinc.xml::1::csteele
-- +goose Up
-- +goose StatementBegin
-- add loinc codes to many tests
UPDATE clinlims.test SET loinc = '6690-2' WHERE name = 'White Blood Cells Count (WBC)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '789-8' WHERE name = 'Red Blood Cells Count (RBC)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '718-7' WHERE name = 'Hemoglobin' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '4544-3' WHERE name = 'Hematocrit' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '787-2' WHERE name = 'Medium corpuscular volum' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '777-3' WHERE name = 'Platelets' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '751-8' WHERE name = 'Neutrophiles' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '770-8' WHERE name = 'Neutrophiles (%)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '731-0' WHERE name = 'Lymphocytes (Abs)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '736-9' WHERE name = 'Lymphocytes (%)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '742-7' WHERE name = 'Monocytes (Abs)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '5905-5' WHERE name = 'Monocytes (%)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '711-2' WHERE name = 'Eosinophiles' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '713-8' WHERE name = 'Eosinophiles (%)' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '704-7' WHERE name = 'Basophiles' AND (loinc is null OR loinc = '');

UPDATE clinlims.test SET loinc = '706-2' WHERE name = 'Basophiles (%)' AND (loinc is null OR loinc = '');
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/loinc.xml::1::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
