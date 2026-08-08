-- source: liquibase liquibase/2.7.x.x/add_tb_observation_type.xml::2023092301::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.observation_history_type(id, type_name, description,lastupdated) VALUES
            (nextval('observation_history_type_seq'),'TbOrderReason','Reason for TB order',now()),
            (nextval('observation_history_type_seq'),'TbDiagnosticReason','Reason for a TB Patient Diagnostic',now()),
            (nextval('observation_history_type_seq'),'TbFollowupReason','Reason for a TB Patient Followup',now()),
            (nextval('observation_history_type_seq'),'TbAnalysisMethod','Method for TB Analysis',now()),
            (nextval('observation_history_type_seq'),'TbSampleAspects','Aspects for a TB sample',now()),
            (nextval('observation_history_type_seq'),'TbFollowupReasonPeriodLine1','Period for a the TB Patient followup 1st Line',now()),
            (nextval('observation_history_type_seq'),'TbFollowupReasonPeriodLine2','Period for a the TB Patient followup 2nd Line',now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_tb_observation_type.xml::2023092301::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
