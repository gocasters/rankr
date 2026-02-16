# Auth Flow Guideline

This document covers the complete authentication and authorization flow in Rankr.

## Flow at a glance

1. `POST /auth/v1/login` issues `access_token` and `refresh_token`.
2. `GET /auth/v1/me` validates access token and returns claims data.
3. `POST /auth/v1/refresh-token` rotates tokens using refresh token.
4. Gateway calls internal `/auth_verify` (mapped to `GET /v1/me`) for protected routes.
5. On success, auth service returns identity headers (`X-User-ID`, `X-Role`, `X-User-Info`) for downstream services.

## Endpoint contract

### Public auth endpoints

- `POST /auth/v1/login`
- `POST /auth/v1/refresh-token`
- `GET /auth/v1/me`
- `GET /auth/v1/health-check`

### `POST /auth/v1/login`

Request:

```bash
curl -X POST http://localhost/auth/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "contributor_name": "fdaei",
    "password": "demo_pass_123"
  }'
```

Response:

```json
{
  "access_token": "<ACCESS_TOKEN>",
  "refresh_token": "<REFRESH_TOKEN>"
}
```

### `GET /auth/v1/me`

Request:

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
```

Response:

```json
{
  "user_id": "<USER_ID>",
  "role": "admin",
  "access": ["*"],
  "expires_at": "<RFC3339>",
  "issued_at": "<RFC3339>"
}
```

Notes:
- Missing `Authorization` header returns `400`.
- Invalid token or invalid role returns `401`.

### `POST /auth/v1/refresh-token`

Refresh token can be sent via:
- `X-Refresh-Token` header (preferred)
- `Refresh-Token` header
- `refresh_token` cookie

Request:

```bash
curl -X POST http://localhost/auth/v1/refresh-token \
  -H "X-Refresh-Token: <REFRESH_TOKEN>"
```

Response:

```json
{
  "access_token": "<NEW_ACCESS_TOKEN>",
  "refresh_token": "<NEW_REFRESH_TOKEN>"
}
```

Notes:
- Missing or invalid refresh token returns `401`.
- Invalid role in token claims returns `401`.

## Setup and full auth flow (dev)

### 1. Start infrastructure

```bash
make infra-up
```

### 2. Start required services

`auth` depends on `contributor` for credential verification.

```bash
make start-contributor-app-dev
make start-auth-app-dev
```

### 3. Verify seeded admin contributor

Migration `000006_seed_admin_contributor.sql` inserts admin contributor `fdaei`.

Check user:

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c \
"SELECT id, github_username, role FROM contributors WHERE github_username = 'fdaei';"
```

If needed, reset password to known value:

```bash
docker exec -it rankr-shared-postgres psql -U contributor_user -d contributor_db -c "
UPDATE contributors
SET password = '\$2a\$10\$gPAOVqKrU6Vtew1eqQU35.XMAhxtIEqo0hiyBOnVWzgh27WwOe0Zq',
    updated_at = NOW()
WHERE github_username = 'fdaei';
"
```

That hash is bcrypt for `demo_pass_123`.

### 4. Login and export tokens

```bash
LOGIN_RESPONSE=$(curl -s -X POST http://localhost/auth/v1/login \
  -H "Content-Type: application/json" \
  -d '{"contributor_name":"fdaei","password":"demo_pass_123"}')

echo "$LOGIN_RESPONSE"
```

Set values manually from response:

```bash
export ACCESS_TOKEN=<ACCESS_TOKEN>
export REFRESH_TOKEN=<REFRESH_TOKEN>
```

### 5. Verify token

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 6. Rotate tokens

```bash
curl -X POST http://localhost/auth/v1/refresh-token \
  -H "X-Refresh-Token: $REFRESH_TOKEN"
```

Update exported tokens with returned values.

## Authorization check flow (`auth_request`)

For protected endpoints, gateway calls internal `/auth_verify`, which proxies to auth service `GET /v1/me` with these headers:

- `Authorization`
- `X-Original-URI`
- `X-Original-Method`
- `X-Original-Host`

Auth service computes required permission and validates against access claims.

Manual permission probe example:

```bash
curl -X GET http://localhost/auth/v1/me \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Original-Method: POST" \
  -H "X-Original-URI: /v1/projects" \
  -H "X-Original-Host: localhost"
```

## Downstream identity headers

When token validation succeeds, auth returns these headers:
- `X-User-ID`
- `X-Role`
- `X-Access`
- `X-User-Info` (base64-encoded user claim)

Other services can read these headers directly or use middleware like `pkg/echo_middleware/require_auth.go`.

## Production variant

Use the same API flow, but start production services:

```bash
make infra-up
make start-contributor-app-prod
make start-auth-app-prod
```

## Quick troubleshooting

- `401 invalid token` on `/me`: check `Authorization: Bearer <token>` format.
- `401 invalid refresh token`: use latest refresh token after each refresh.
- `login` fails: ensure contributor service is running and seeded user exists.
- protected routes always fail: verify gateway config includes `auth_request /auth_verify` in route block.
