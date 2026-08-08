-- source: liquibase liquibase/2.7.x.x/add_recency_testing.xml::202309102::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization(
            id, description, english,
            french, lastupdated)
            VALUES
            (nextval('localization_seq'),'test name','Asante HIV-1 Rapid Recency','Asante HIV-1 Rapid Recency',
            now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(
            id, description, english,
            french, lastupdated)
            VALUES
            (nextval('localization_seq'),'test report name','Rapid Test for Recent infection','Test Rapide pour Infection Récente',
            now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_recency_testing.xml::202309102::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
