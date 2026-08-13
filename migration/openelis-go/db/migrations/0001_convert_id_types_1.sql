-- source: liquibase liquibase/2.0.x.x/convert_id_types.xml::1::caleb
-- +goose Up
-- +goose StatementBegin
-- modify login_user as an example
ALTER TABLE clinlims.login_user ALTER COLUMN id TYPE INTEGER USING (id::INTEGER);
DO $$ DECLARE constraint_name varchar;
BEGIN
  SELECT tc.CONSTRAINT_NAME into strict constraint_name
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
    WHERE CONSTRAINT_TYPE = 'PRIMARY KEY'
      AND TABLE_NAME = 'login_user' AND TABLE_SCHEMA = 'clinlims';
    EXECUTE 'alter table clinlims.login_user drop constraint "' || constraint_name || '"';
END $$;
ALTER TABLE clinlims.login_user ADD PRIMARY KEY (id);
ALTER TABLE clinlims.login_user ADD IF NOT EXISTS last_updated TIMESTAMP WITH TIME ZONE;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.0.x.x/convert_id_types.xml::1::caleb
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
