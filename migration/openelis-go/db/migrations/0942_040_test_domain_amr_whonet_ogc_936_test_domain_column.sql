-- source: liquibase liquibase/3.5.x.x/040-test-domain-amr-whonet.xml::OGC-936-test-domain-column::OGC
-- +goose Up
-- +goose StatementBegin
ALTER TABLE clinlims.test ADD IF NOT EXISTS domain VARCHAR(20) DEFAULT 'CLINICAL' NOT NULL;
ALTER TABLE clinlims.test
            ADD CONSTRAINT test_domain_chk
            CHECK (domain IN ('CLINICAL', 'ENVIRONMENTAL', 'VECTOR'));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/040-test-domain-amr-whonet.xml::OGC-936-test-domain-column::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
