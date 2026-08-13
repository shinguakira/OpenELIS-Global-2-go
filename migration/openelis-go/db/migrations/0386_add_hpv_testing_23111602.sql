-- source: liquibase liquibase/2.7.x.x/add_hpv_testing.xml::23111602::CIV developer Group
-- +goose Up
-- +goose StatementBegin
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV 16','HPV 16',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV 16','HPV 16',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV 18','HPV 18',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV 18','HPV 18',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','Other HR HPV','Autre HPV HR',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','Other HR HPV','Autre HPV HR',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV 18_45','HPV 18_45',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV 18_45','HPV 18_45',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV P3','HPV P3',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV P3','HPV P3',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV P4','HPV P4',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV P4','HPV P4',now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES	(nextval('localization_seq'),'test name','HPV P5','HPV P5',	now()) ON CONFLICT DO NOTHING;
INSERT INTO clinlims.localization(id, description, english,	french, lastupdated)
            VALUES (nextval('localization_seq'),'test report name','HPV P5','HPV P5',now()) ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- TODO: no safe auto-generated rollback for this changeset.
-- Liquibase source: liquibase/2.7.x.x/add_hpv_testing.xml::23111602::CIV developer Group
-- Hand-write if this migration must be reversible; see
-- migration/liquibase-to-goose-plan.md sec 7 (Risk items).
