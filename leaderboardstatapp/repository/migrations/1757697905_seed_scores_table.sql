-- +migrate Up
INSERT into scores (contributor_id, activity , score )
VALUES (1, 'pr', 3),
       (2, 'pr', 3),
       (4, 'pr', 3),
       (4, 'review', 2),
       (4, 'review', 2),
       (4, 'review', 2),
       (5, 'pr', 3),
       (5, 'pr', 3),
       (5, 'review', 3),
       (5, 'review', 3);
