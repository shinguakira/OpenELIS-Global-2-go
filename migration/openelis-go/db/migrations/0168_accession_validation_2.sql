-- source: liquibase liquibase/2.3.x.x/accession_validation.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
-- Insert in Validation by search by accession
INSERT INTO clinlims.menu(id, parent_id, presentation_order, element_id,
            action_url, click_action,
            display_key, tool_tip_key, new_window, is_active)
            VALUES
            (nextval('clinlims.menu_seq'),(select id from clinlims.menu where
            element_id='menu_resultvalidation'),20,'menu_accession_validation','/AccessionValidation.do',default,'menu.accession.validation','tooltip.accession.validation',default,default) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/accession_validation.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
