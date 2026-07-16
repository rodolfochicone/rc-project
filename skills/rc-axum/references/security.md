# Axum — security

## Network posture (VPS)

- Listen on **127.0.0.1**; terminate TLS at Caddy/Nginx.
- Do not open API ports (3000) on the public firewall.
- Trust `X-Forwarded-For` only from the local proxy (configure carefully if used).

## CORS

- Dev may use loose CORS; **prod: explicit `AllowOrigin` list**, methods, headers.
- Credentials + `*` origin is invalid — set concrete origins when cookies are involved.
- Prefer same-origin via reverse proxy (`/api` → Axum) so the browser needs no CORS for first-party apps.

## Authentication & authorization

- Validate bearer/session **before** sensitive handlers and **before** WebSocket upgrade.
- Put identity in request extensions after middleware; handlers must not re-parse tokens ad hoc in inconsistent ways.
- Use constant-time compares for secrets where applicable; store password hashes with a modern KDF (Argon2), never plain.
- Short-lived access tokens; rotate secrets; never log Authorization headers.

## Input validation

- Deserialize with serde; reject unknown critical fields when needed (`deny_unknown_fields` on sensitive DTOs).
- Enforce size limits on bodies (`DefaultBodyLimit` / tower-http limit layers).
- Path/UUID parsing failures → 400, not 500.

## Headers & responses

- Do not reflect unsanitized input into headers.
- Avoid detailed internal errors in production JSON (`AppError::Internal` → generic message + log).
- Set security headers at the **proxy** (HSTS, frame deny) and/or middleware if not proxy-owned.

## Secrets & config

- Load `DATABASE_URL`, JWT keys, etc. from env; never commit `.env`.
- Fail fast on missing required secrets at startup.

## Dependencies

- Run `cargo audit` / `cargo deny` in CI when available.
- Keep `axum` / `hyper` / `tokio` within supported minor lines.

## WebSocket-specific

- Authenticate on HTTP upgrade request (cookie or `Sec-WebSocket-Protocol` / query only if unavoidable — prefer cookie/header via same-origin).
- Validate message schema; reject oversized text/binary.
- Rate-limit message volume per connection.

## Checklist before ship

- [ ] Bound to loopback or intentional public interface
- [ ] CORS locked down or same-origin proxy
- [ ] Auth on mutating and private read routes + WS
- [ ] Body size limits
- [ ] No secrets in logs/traces
- [ ] Dependency audit in CI
