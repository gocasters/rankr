-- +migrate Up
INSERT INTO permissions (name, description)
VALUES
    ('*', 'all permissions'),
    ('contributor:read', 'read contributor data'),
    ('contributor:update', 'update contributor data'),
    ('score:read', 'read scores'),
    ('project:read', 'read projects'),
    ('contributor:create', 'create contributors')
ON CONFLICT (name) DO NOTHING;

INSERT INTO roles (name, description)
VALUES
    ('admin', 'full access'),
    ('user', 'standard access')
ON CONFLICT (name) DO NOTHING;

WITH admin_role AS (
    SELECT id FROM roles WHERE name = 'admin'
),
perms AS (
    SELECT id FROM permissions
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT admin_role.id, perms.id
FROM admin_role, perms
ON CONFLICT (role_id, permission_id) DO NOTHING;

WITH user_role AS (
    SELECT id FROM roles WHERE name = 'user'
),
perms AS (
    SELECT id FROM permissions WHERE name NOT IN (
        'contributor:create',
        '*'
    )
)
INSERT INTO role_permissions (role_id, permission_id)
SELECT user_role.id, perms.id
FROM user_role, perms
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- +migrate Down
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('admin', 'user'))
  AND permission_id IN (
      SELECT id FROM permissions WHERE name IN (
          '*',
          'contributor:read',
          'contributor:update',
          'score:read',
          'project:read',
          'contributor:create'
      )
  );

DELETE FROM roles WHERE name IN ('admin', 'user');

DELETE FROM permissions WHERE name IN (
    '*',
    'contributor:read',
    'contributor:update',
    'score:read',
    'project:read',
    'contributor:create'
);
