-- source: liquibase liquibase/2.3.x.x/results_range.xml::2::csteele
-- +goose Up
-- +goose StatementBegin
-- Insert in Results by search by accession range
INSERT INTO clinlims.menu(id, parent_id, presentation_order, element_id,
            action_url, click_action,
            display_key, tool_tip_key, new_window, is_active)
            VALUES
            (nextval('clinlims.menu_seq'),(select id from clinlims.menu where
            element_id='menu_results_search'),3,'menu_results_range','/RangeResults.do',default,'menu.results.range','tooltip.results.range',default,default) ON CONFLICT DO NOTHING;
UPDATE clinlims.menu SET presentation_order = 4 WHERE element_id = 'menu_results_status';
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.3.x.x/results_range.xml::2::csteele
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
