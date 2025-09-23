-- +migrate Up
INSERT into scores (contributor_id, activity , score )
VALUES (10, 'fix', 1);

