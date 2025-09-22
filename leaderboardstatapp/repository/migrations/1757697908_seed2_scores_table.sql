-- +migrate Up
INSERT into scores (contributor_id, activity , score )
VALUES (9, 'fix', 1);

