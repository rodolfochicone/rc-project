# SQLx / Postgres — security

## SQL injection

- **Only** pass dynamic values via `.bind` / macro arguments.
- Never build SQL with `format!`, `+`, or string replace from user input.
- Dynamic **identifiers** (sort column names) must be allowlisted in Rust, not concatenated from raw input.

```rust
fn order_clause(sort: &str) -> &'static str {
    match sort {
        "created_at" => "created_at",
        "name" => "name",
        _ => "id",
    }
}
```

## Database roles

- App role: `CONNECT` + DML on app schema only; no `SUPERUSER`, no `CREATEDB` in prod.
- Migrations: separate role or controlled CI job.
- Revoke public grants on sensitive tables.

## Secrets

- `DATABASE_URL` from env/secret manager; never in git.
- Prefer SCRAM auth; require SSL to remote Postgres (`sslmode=require` when not local docker).

## Data exposure

- Do not `SELECT *` into API responses — project columns.
- Redact PII in logs; log query **names**/error codes, not full bound password fields.

## Multi-tenant

- Enforce `tenant_id` (or RLS) on every query; never trust client-sent tenant without auth binding.

## Connection limits

- Cap pool size so app × instances does not exhaust Postgres `max_connections`.
- On shared VPS Postgres, leave room for admin and backups.

## Checklist

- [ ] All dynamic SQL uses binds
- [ ] Identifier allowlists where needed
- [ ] Least-privilege role
- [ ] SSL for non-local DB
- [ ] No secrets in repo or traces
- [ ] Migrations reviewed for destructive ops
