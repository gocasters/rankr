## Flow

- Use seeded admin contributor created by migration `000006_seed_admin_contributor.sql`.
- Login from `AuthService` (`POST /auth/v1/login`) and get JWT tokens.
- Validate and rotate token with `GET /auth/v1/me`.
- For other modules, OpenResty calls internal `/auth_verify` on every protected request.
- If token is valid, these headers are forwarded to upstream services: `X-User-ID`, `X-Role`, `X-User-Info`.

> **Note**:
> - Public auth endpoints are under `/auth/*` in gateway.
> - Gateway route selection is path-based; `Host` header is not required.
> - Protected routes in other modules are checked by `auth_request /auth_verify`.

---

## Steps

### 1. Start infrastructure and required services (prod)

```bash
make infra-up
make start-contributor-app-prod
make start-auth-app-prod
make start-project-app-prod
```

### 2. Verify seeded admin user

Migration `000006_seed_admin_contributor.sql` inserts admin user `fdaei`.

Check user exists:

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c \
"SELECT id, github_username, role FROM contributors WHERE github_username = 'fdaei';"
```

If you do not know the current password, set a known one:

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c "
UPDATE contributors
SET password = '\$2a\$10\$gPAOVqKrU6Vtew1eqQU35.XMAhxtIEqo0hiyBOnVWzgh27WwOe0Zq',
    updated_at = NOW()
WHERE github_username = 'fdaei';
"
```

`$2a$10$gPAOVqKrU6Vtew1eqQU35.XMAhxtIEqo0hiyBOnVWzgh27WwOe0Zq` is bcrypt hash for `demo_pass_123`.

### 3. Login with seeded admin

```bash
curl -X POST http://localhost/auth/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "contributor_name": "fdaei",
    "password": "demo_pass_123"
  }'
```

Example response:

```json
{
  "access_token": "<ACCESS_TOKEN>",
  "refresh_token": "<REFRESH_TOKEN>"
}
```

Set admin token for next steps:

```bash
export ADMIN_ACCESS_TOKEN=<ACCESS_TOKEN>
```

### 4. Verify token explicitly

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInJvbGUiOiJhZG1pbiIsImFjY2VzcyI6WyIqIl0sImV4cCI6MTc3MTE3NjgzMCwiaWF0IjoxNzcxMTczMjMwfQ.wVAc6Wxc8WiR3TGGgVAh-iXwSX1K5EXYYjIQEiran6M"
```

Expected response (example):

```json
{
  "user_id": "<USER_ID>",
  "role": "admin",
  "access": [
    "*"
  ],
  "access_token": "<NEW_ACCESS_TOKEN>",
  "refresh_token": "<NEW_REFRESH_TOKEN>",
  "expires_at": "<RFC3339>",
  "issued_at": "<RFC3339>"
}
```

### 5. Create a new user

Create a new contributor via ContributorService (protected route, use admin token):

```bash
curl -X POST http://localhost/v1/create \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "github_id": 1001001,
    "github_username": "new_user_demo",
    "password": "demo_pass_123",
    "display_name": "New User",
    "privacy_mode": "real"
  }'
```

Expected response:

```json
{
  "id": 2
}
```

Check user record:

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c \
"SELECT id, github_username, role FROM contributors WHERE github_username = 'new_user_demo';"
```

By default, created users get role `user`.

### 6. Login with the new user and inspect permissions

```bash
curl -X POST http://localhost/auth/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "contributor_name": "new_user_demo",
    "password": "demo_pass_123"
  }'
```

Set the returned token:

```bash
export USER_ACCESS_TOKEN=<ACCESS_TOKEN>
```

Inspect role and access list:

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer $USER_ACCESS_TOKEN"
```

Expected behavior:
- `role` should be `user`
- `access` should include user-level permissions from RBAC seed (for example `project:read`, `contributor:read`, `contributor:update`, ...)

### 7. Optional: promote user to admin and re-check access

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c "
UPDATE contributors
SET role = 'admin',
    updated_at = NOW()
WHERE github_username = 'new_user_demo';
"
```

Login again and verify token. Expected behavior:
- `role` becomes `admin`
- `access` includes `*` (full access)

### 8. Token check for protected routes (project module)

Without token (should fail):

```bash
curl -i http://localhost/v1/projects
```

With token (should pass auth check):

```bash
curl -i http://localhost/v1/projects \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

### 9. Optional permission check

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN" \
  -H "X-Original-Method: POST" \
  -H "X-Original-URI: /v1/projects" \
  -H "X-Original-Host: localhost"
```

For `admin` token, this request should pass (not forbidden).

```json
{
  "user_id": "<USER_ID>",
  "role": "admin",
  "access": ["*"]
}
```

### 10. Verify token check is enabled in the default gateway

```bash
rg -n "auth_request /auth_verify;" deploy/infrastructure/openresty/development/conf.d/auth.conf
```

You should see matches in route blocks for service paths in:
- `deploy/infrastructure/openresty/development/conf.d/auth.conf`
