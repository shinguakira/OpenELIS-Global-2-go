-- source: liquibase liquibase/3.3.x.x/032-label-preset-is-universal.xml::label-preset-is-universal-002::ogc-285
-- +goose Up
-- +goose StatementBegin
-- FR-014a: mark the seeded Specimen Label system preset universal — every sample gets a specimen label; tests only override qty
UPDATE clinlims.label_preset
             SET is_universal = true, last_updated = CURRENT_TIMESTAMP
             WHERE is_system = true AND name = 'Specimen Label';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/032-label-preset-is-universal.xml::label-preset-is-universal-002::ogc-285
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
