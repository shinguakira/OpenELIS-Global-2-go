-- source: liquibase liquibase/3.3.x.x/028-create-site-branding-table.xml::028-create-site-branding-table::openelis
-- +goose Up
-- +goose StatementBegin
-- Creating site_branding table for white labeling / site branding feature
CREATE SEQUENCE  IF NOT EXISTS site_branding_seq START WITH 1 INCREMENT BY 1;
CREATE TABLE IF NOT EXISTS site_branding (id INTEGER DEFAULT nextval('site_branding_seq') NOT NULL, header_logo_path VARCHAR(500), login_logo_path VARCHAR(500), use_header_logo_for_login BOOLEAN DEFAULT FALSE NOT NULL, favicon_path VARCHAR(500), header_color VARCHAR(50) DEFAULT '''#295785''' NOT NULL, primary_color VARCHAR(50) DEFAULT '''#0f62fe''' NOT NULL, secondary_color VARCHAR(50) DEFAULT '''#393939''' NOT NULL, color_mode VARCHAR(10) DEFAULT '''light''' NOT NULL, sys_user_id VARCHAR(255) NOT NULL, last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL, CONSTRAINT site_branding_pkey PRIMARY KEY (id));
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/3.3.x.x/028-create-site-branding-table.xml::028-create-site-branding-table::openelis
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
