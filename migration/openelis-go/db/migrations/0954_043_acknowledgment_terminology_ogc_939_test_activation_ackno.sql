-- source: liquibase liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-test-activation-acknowledgment-table::OGC
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clinlims.test_activation_acknowledgment (id VARCHAR(36) NOT NULL, test_id numeric(10) NOT NULL, user_id numeric(10) NOT NULL, acknowledged_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, gaps_acknowledged JSONB, lastupdated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE, CONSTRAINT pk_test_activation_acknowledgment PRIMARY KEY (id));
ALTER TABLE clinlims.test_activation_acknowledgment ADD CONSTRAINT fk_test_activation_ack_test FOREIGN KEY (test_id) REFERENCES clinlims.test (id);
ALTER TABLE clinlims.test_activation_acknowledgment ADD CONSTRAINT fk_test_activation_ack_user FOREIGN KEY (user_id) REFERENCES clinlims.system_user (id);
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.5.x.x/043-acknowledgment-terminology.xml::OGC-939-test-activation-acknowledgment-table::OGC
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
