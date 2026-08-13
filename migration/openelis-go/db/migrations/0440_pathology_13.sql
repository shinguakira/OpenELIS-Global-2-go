-- source: liquibase liquibase/2.8.x.x/pathology.xml::13::csteele
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.program (id, lastupdated, code, name) VALUES (nextval('clinlims.program_seq'), NOW(), 'ROUTINE', 'Routine Testing') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.program (id, lastupdated, code, name) VALUES (nextval('clinlims.program_seq'), NOW(), 'HIV_INIT', 'People living with HIV Program - Initial Visit') ON CONFLICT DO NOTHING;
INSERT INTO clinlims.program (id, lastupdated, code, name) VALUES (nextval('clinlims.program_seq'), NOW(), 'HIV_FOLLOW', 'People living with HIV Program - Follow-up Visit') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.8.x.x/pathology.xml::13::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
