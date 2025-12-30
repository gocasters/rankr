-- +migrate Up
INSERT INTO contributors (
    github_id,
    github_username,
    email,
    password,
    role,
    privacy_mode,
    created_at,
    updated_at
) VALUES (
    NULL,
    'fdaei',
    NULL,
    '123456789',
    'admin',
    'real',
    NOW(),
    NOW()
)
ON CONFLICT (github_username)
DO UPDATE SET role = EXCLUDED.role;

-- +migrate Down
DELETE FROM contributors WHERE github_username = 'fdaei';
