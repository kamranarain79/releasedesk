# Infrastructure

ReleaseDesk keeps **metadata** in a database and **firmware bytes** on disk. Do not encrypt or recompress `.bin`, `.hex`, or `.raucb` inside the app. Factory flashing and Filebrowser need the original bytes. Encrypt the disk (LUKS or the host volume) and the Postgres data directory instead.

## Database

| Mode | When |
|------|------|
| PostgreSQL | `docker compose up`. Compose starts Postgres next to the app and sets `RELEASEDESK_DATABASE_URL`. The database port is not published. |
| SQLite file `_data/releasedesk.db` | The binary on a host, when `RELEASEDESK_DATABASE_URL` is unset. |

```bash
docker compose up --build
```

Postgres is not published on the host. Only ReleaseDesk talks to it. Change the Compose password before exposing the console on a public network.

Backup: `pg_dump` plus a copy of the `storage/` directory under `RELEASEDESK_DATA_DIR`. Restore both together. There is no automatic SQLite-to-Postgres copy.

## Process

- Console `:8080`, portal `:9090`, HTTP inside the process.
- Put TLS on a reverse proxy. Set `RELEASEDESK_COOKIE_SECURE=1` and `RELEASEDESK_JWT_SECRET` to a long random value.
- `RELEASEDESK_ENV=production` refuses to boot if the JWT secret is missing or still `secret`.
- `RELEASEDESK_MAX_UPLOAD_MB` caps uploads (default 512).
- Artifact paths are `storage/{sha256[0:2]}/{sha256}` with mode `0750` directories. Downloads accept only 64 hex characters.

## Tokens and Teams

Mint a CI token on the app Settings page. Store it in GitLab as a masked, protected variable. The instance token is generated on first start and shown once under Administration. `RELEASEDESK_API_TOKEN`, when set, is accepted as well.

Teams incoming webhooks must be `https` on `outlook.office.com` or `*.webhook.office.com`. A failed webhook does not fail the CI request.

## Filebrowser

Leave Filebrowser as its own service. The Python tool lists `/api/v1/.../files` or downloads `bundle.zip` (includes `sha256sums.txt`), then copies files into Filebrowser for flashing.

Outgoing mail is configured on the Administration page (host, port, user, password, from). `SMTP_*` environment variables are copied in only when that form has never been saved.
