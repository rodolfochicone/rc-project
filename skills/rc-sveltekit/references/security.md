# SvelteKit — security

## CSRF

- SvelteKit **origin-checks** form POSTs by default — keep that enabled.
- For cross-site API use explicit CORS on Axum; do not weaken Kit CSRF for convenience.

## Cookies & sessions

```ts
cookies.set('sessionid', value, {
	path: '/',
	httpOnly: true,
	secure: true, // HTTPS prod
	sameSite: 'lax', // or 'strict' if UX allows
	maxAge: 60 * 60 * 24 * 7
});
```

- Prefer **httpOnly** session cookies over storing tokens in `localStorage`.
- Clear cookies on logout with matching `path`.
- After logout/login in an action, set `event.locals.user` accordingly.

## XSS

- Prefer text interpolation; avoid `{@html}` with untrusted content.
- If `{@html}` is required, sanitize with a maintained library and strict policy.
- CSP via `kit.csp` in `svelte.config.js` (nonce/hash modes) reduces inline script risk.

## CSP (sketch)

```js
// svelte.config.js — conceptual
kit: {
  csp: {
    mode: 'auto',
    directives: {
      'default-src': ['self'],
      'connect-src': ['self'], // add wss:// if needed
      'img-src': ['self', 'data:']
    }
  }
}
```

Tune for analytics/fonts; prefer tightening over `unsafe-inline`.

## Secrets

- Never import server modules that read private env from client components.
- Put server-only code under `$lib/server` (or ensure no client import chain).
- `PUBLIC_*` is always visible — no API keys.

## AuthZ

- Check `locals.user` in `+page.server.ts` / actions; `redirect(303, '/login')` when missing.
- Do not rely on hiding links in the UI as authorization.

## Proxy / SSR SSRF

- `load` that fetches user-controlled URLs can become SSRF — allowlist hosts.
- Prefer fixed `API_INTERNAL_URL` for the Axum service.

## Checklist

- [ ] CSRF left on
- [ ] Session cookies httpOnly + secure + sameSite
- [ ] No secrets in PUBLIC_ or client bundles
- [ ] Authorization in server loads/actions
- [ ] CSP considered for production
- [ ] `{@html}` audited or absent
