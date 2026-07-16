# SvelteKit — testing

## Layers

| Layer | Tool | What |
| ----- | ---- | ---- |
| Unit | Vitest | pure utils, runes-free logic |
| Component | Vitest + @testing-library/svelte | UI behavior |
| Server load/actions | Vitest with mocked `fetch`/cookies | validation branches |
| E2E | Playwright | SSR + forms + auth cookies |

Use `rc-vitest` for Vitest API depth.

## Unit example

```ts
import { describe, it, expect } from 'vitest';
import { publicApiUrl } from '$lib/config';

describe('publicApiUrl', () => {
	it('falls back locally', () => {
		expect(publicApiUrl()).toMatch(/127\.0\.0\.1|localhost/);
	});
});
```

## Testing form actions

- Call the action function with a mock `Request` + `cookies` API.
- Assert `fail(400, …)` payloads and cookie `set` calls.

## E2E (Playwright sketch)

```ts
import { test, expect } from '@playwright/test';

test('home SSR shows API hello or fallback', async ({ page }) => {
	await page.goto('/');
	await expect(page.locator('h1')).toBeVisible();
});

test('dashboard connects or shows error UI', async ({ page }) => {
	await page.goto('/dashboard');
	await expect(page.getByText(/Dashboard|WebSocket|conectado|desconectado/i)).toBeVisible();
});
```

## Typecheck

```bash
bun run check          # svelte-check
bun run build          # fails on hard errors
```

## Lint

```bash
bun run lint           # eslint/prettier if configured
```

## CI minimum

```bash
bun install --frozen-lockfile
bun run check
bun run test           # if present
bun run build
```

## Anti-patterns

- Only testing client navigation without SSR (misses `+page.server` failures)
- E2E against production credentials
- Snapshotting entire HTML without asserting user-visible contracts
