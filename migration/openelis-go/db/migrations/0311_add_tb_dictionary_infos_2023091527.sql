-- source: liquibase liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091527::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.dictionary_category (id, description, lastupdated, local_abbrev, name) VALUES (nextval('clinlims.dictionary_category_seq'), 'TB Followup Reasons', NOW(), 'TBFReason', 'TB Followup Reasons') ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_dictionary_infos.xml::2023091527::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
