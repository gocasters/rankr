## Flow

- Use seeded admin contributor created by migration `000006_seed_admin_contributor.sql`.
- Login from `AuthService` (`POST /auth/v1/login`) and get JWT tokens.
- Validate token with `POST /auth/v1/token/verify`.
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

### 4. Verify token explicitly

```bash
curl -X POST http://localhost/auth/v1/token/verify \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInJvbGUiOiJhZG1pbiIsImFjY2VzcyI6WyIqIl0sImV4cCI6MTc3MTE1MTczMSwiaWF0IjoxNzcxMTQ4MTMxfQ.XmdFQeXFEU4hoi4ohwdySzdrs0BM3OaclTG2l3_xuEA"
```

Expected response (example):

```json
{
  "user_id": "<USER_ID>",
  "role": "admin",
  "access": [
    "*"
  ],
  "expires_at": "<RFC3339>",
  "issued_at": "<RFC3339>"
}
```

### 5. Token check for protected routes (project module)

Without token (should fail):

```bash
curl -i http://localhost/v1/projects
```

With token (should pass auth check):

```bash
curl -i http://localhost/v1/projects \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInJvbGUiOiJhZG1pbiIsImFjY2VzcyI6WyIqIl0sImV4cCI6MTc3MTE1MTczMSwiaWF0IjoxNzcxMTQ4MTMxfQ.XmdFQeXFEU4hoi4ohwdySzdrs0BM3OaclTG2l3_xuEA"
```

### 6. Optional permission check

```bash
curl -X POST http://localhost/auth/v1/token/verify \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMSIsInJvbGUiOiJhZG1pbiIsImFjY2VzcyI6WyIqIl0sImV4cCI6MTc3MTE1MTczMSwiaWF0IjoxNzcxMTQ4MTMxfQ.XmdFQeXFEU4hoi4ohwdySzdrs0BM3OaclTG2l3_xuEA" \
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

### 7. Verify token check is enabled in the default gateway

```bash
rg -n "auth_request /auth_verify;" deploy/infrastructure/openresty/development/conf.d/auth.conf
```

You should see matches in route blocks for service paths in:
- `deploy/infrastructure/openresty/development/conf.d/auth.conf`
