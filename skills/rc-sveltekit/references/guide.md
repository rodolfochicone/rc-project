# SvelteKit 2 — guide

Sources: official SvelteKit docs (sveltejs/kit). Svelte 5 runes syntax for UI.

## Project shape

```text
src/
  app.d.ts
  app.html
  hooks.server.ts
  lib/
    server/          # server-only modules
    config.ts        # public helpers only
  routes/
    +layout.svelte
    +page.server.ts
    +page.svelte
    dashboard/
      +page.svelte
```

## Server load (SSR)

```ts
// src/routes/+page.server.ts
import type { PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';

export const load: PageServerLoad = async ({ fetch }) => {
	const base = env.API_INTERNAL_URL ?? 'http://127.0.0.1:3000';
	const res = await fetch(`${base}/api/hello`);
	const hello = res.ok ? await res.json() : null;
	return { hello };
};
```

- Use **internal** API URL on the server (loopback), not the public browser origin, when both run on one VPS.
- `fetch` in load is instrumented by Kit (cookies, origin) for same-app requests; for external Axum, absolute URL is fine.

## Page (Svelte 5)

```svelte
<script lang="ts">
	let { data } = $props();
</script>

<svelte:head>
	<title>Home</title>
	<meta name="description" content="SSR page" />
</svelte:head>

<h1>{data.hello?.message ?? '…'}</h1>
```

## Form actions

```ts
// +page.server.ts
import { fail } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
	login: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = data.get('email');
		if (!email || typeof email !== 'string') {
			return fail(400, { email, missing: true });
		}
		// verify user...
		cookies.set('sessionid', '…', {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure: true // prod HTTPS
		});
		return { success: true };
	}
};
```

```svelte
<form method="POST" action="?/login">
	<input name="email" type="email" />
	<button type="submit">Login</button>
</form>
```

## hooks.server + locals

```ts
// src/hooks.server.ts
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	const sid = event.cookies.get('sessionid');
	event.locals.user = sid ? await getUser(sid) : null;
	return resolve(event);
};
```

```ts
// src/app.d.ts
declare global {
	namespace App {
		interface Locals {
			user: { name: string } | null;
		}
	}
}
export {};
```

**Important:** `handle` does not re-run between an action and the following `load` on the same request. If an action changes cookies/session, **update `event.locals` in the action** so loads see fresh state.

## Env vars

| Prefix / API | Visibility |
| ------------ | ---------- |
| `$env/static/public` `PUBLIC_*` | Client + server |
| `$env/dynamic/private` | Server only |
| `API_INTERNAL_URL` | Server SSR → Axum |

## Client WebSocket (dashboard)

```ts
const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const ws = new WebSocket(`${proto}//${window.location.host}/ws/metrics`);
```

Proxy `/ws/*` to Axum at the reverse proxy so the browser stays same-origin.

## Layout data

`+layout.server.ts` for shared session/nav data; avoid N+1 duplicate fetches in every page.

## Streaming / defer

Use Kit streaming patterns when slow secondary data should not block TTFB for critical SEO content.
