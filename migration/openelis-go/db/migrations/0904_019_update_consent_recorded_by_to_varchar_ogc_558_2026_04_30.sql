-- source: liquibase liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-3::herbertYiga
-- +goose Up
-- +goose StatementBegin
UPDATE clinlims.sample s
            SET consent_recorded_by = (
                SELECT TRIM(BOTH FROM (su.first_name || ' ' || su.last_name))
                FROM clinlims.system_user su
                WHERE CAST(su.id AS BIGINT) = CAST(s.consent_recorded_by AS BIGINT)
            )
            WHERE s.consent_recorded_by IS NOT NULL
              AND s.consent_recorded_by ~ '^[0-9]+$';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/019-update-consent-recorded-by-to-varchar.xml::OGC-558-2026-04-30-3::herbertYiga
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
