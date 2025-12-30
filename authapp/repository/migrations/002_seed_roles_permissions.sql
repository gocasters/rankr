-- +migrate Up
INSERT INTO permissions (name, description)
VALUES
    ('*', 'all permissions'),
    ('argus:read', 'read argus data'),
    ('argus:create', 'create argus data'),
    ('argus:update', 'update argus data'),
    ('argus:delete', 'delete argus data'),
    ('contributor:read', 'read contributor data'),
    ('contributor:create', 'create contributor data'),
    ('contributor:update', 'update contributor data'),
    ('contributor:delete', 'delete contributor data'),
    ('leaderboardscoring:read', 'read leaderboard scoring data'),
    ('leaderboardscoring:create', 'create leaderboard scoring data'),
    ('leaderboardscoring:update', 'update leaderboard scoring data'),
    ('leaderboardscoring:delete', 'delete leaderboard scoring data'),
    ('leaderboardstat:read', 'read leaderboard stats data'),
    ('leaderboardstat:create', 'create leaderboard stats data'),
    ('leaderboardstat:update', 'update leaderboard stats data'),
    ('leaderboardstat:delete', 'delete leaderboard stats data'),
    ('notification:read', 'read notifications'),
    ('notification:create', 'create notifications'),
    ('notification:update', 'update notifications'),
    ('notification:delete', 'delete notifications'),
    ('project:read', 'read projects'),
    ('project:create', 'create projects'),
    ('project:update', 'update projects'),
    ('project:delete', 'delete projects'),
    ('realtime:read', 'read realtime data'),
    ('realtime:create', 'create realtime data'),
    ('realtime:update', 'update realtime data'),
    ('realtime:delete', 'delete realtime data'),
    ('task:read', 'read tasks'),
    ('task:create', 'create tasks'),
    ('task:update', 'update tasks'),
    ('task:delete', 'delete tasks'),
    ('userprofile:read', 'read user profiles'),
    ('userprofile:create', 'create user profiles'),
    ('userprofile:update', 'update user profiles'),
    ('userprofile:delete', 'delete user profiles'),
    ('webhook:read', 'read webhooks'),
    ('webhook:create', 'create webhooks'),
    ('webhook:update', 'update webhooks'),
    ('webhook:delete', 'delete webhooks')
ON CONFLICT (name) DO NOTHING;

WITH admin_role AS (
    SELECT id FROM permissions WHERE name = '*'
)
INSERT INTO role_permissions (role, permission_id)
SELECT 'admin'::role_enum, admin_role.id
FROM admin_role
ON CONFLICT (role, permission_id) DO NOTHING;

WITH user_perms AS (
    SELECT id FROM permissions WHERE name IN (
        'argus:read',
        'contributor:read',
        'contributor:update',
        'leaderboardscoring:read',
        'leaderboardstat:read',
        'notification:read',
        'notification:update',
        'project:read',
        'realtime:read',
        'task:read',
        'userprofile:read'
    )
)
INSERT INTO role_permissions (role, permission_id)
SELECT 'user'::role_enum, user_perms.id
FROM user_perms
ON CONFLICT (role, permission_id) DO NOTHING;

-- +migrate Down
DELETE FROM role_permissions
WHERE role IN ('admin', 'user')
  AND permission_id IN (
      SELECT id FROM permissions WHERE name IN (
          '*',
          'argus:read',
          'argus:create',
          'argus:update',
          'argus:delete',
          'contributor:read',
          'contributor:create',
          'contributor:update',
          'contributor:delete',
          'leaderboardscoring:read',
          'leaderboardscoring:create',
          'leaderboardscoring:update',
          'leaderboardscoring:delete',
          'leaderboardstat:read',
          'leaderboardstat:create',
          'leaderboardstat:update',
          'leaderboardstat:delete',
          'notification:read',
          'notification:create',
          'notification:update',
          'notification:delete',
          'project:read',
          'project:create',
          'project:update',
          'project:delete',
          'realtime:read',
          'realtime:create',
          'realtime:update',
          'realtime:delete',
          'task:read',
          'task:create',
          'task:update',
          'task:delete',
          'userprofile:read',
          'userprofile:create',
          'userprofile:update',
          'userprofile:delete',
          'webhook:read',
          'webhook:create',
          'webhook:update',
          'webhook:delete'
      )
  );

DELETE FROM permissions WHERE name IN (
    '*',
    'argus:read',
    'argus:create',
    'argus:update',
    'argus:delete',
    'contributor:read',
    'contributor:create',
    'contributor:update',
    'contributor:delete',
    'leaderboardscoring:read',
    'leaderboardscoring:create',
    'leaderboardscoring:update',
    'leaderboardscoring:delete',
    'leaderboardstat:read',
    'leaderboardstat:create',
    'leaderboardstat:update',
    'leaderboardstat:delete',
    'notification:read',
    'notification:create',
    'notification:update',
    'notification:delete',
    'project:read',
    'project:create',
    'project:update',
    'project:delete',
    'realtime:read',
    'realtime:create',
    'realtime:update',
    'realtime:delete',
    'task:read',
    'task:create',
    'task:update',
    'task:delete',
    'userprofile:read',
    'userprofile:create',
    'userprofile:update',
    'userprofile:delete',
    'webhook:read',
    'webhook:create',
    'webhook:update',
    'webhook:delete'
);
