-- +migrate Up
INSERT into scores (contributor_id, project_id, activity, score, earned_at )
VALUES (8,5, 'fix', 1, '2025-1-09 12:38:34.943018+00');

