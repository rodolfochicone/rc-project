---
name: rc-tanstack
description: Comprehensive guide for TanStack ecosystem in React - Query/DB best practices for data fetching, caching, mutations and server state, Form for form handling, Router best practices for type-safe routing, data loading, search params and navigation, and Start best practices for full-stack apps (server functions, middleware, SSR, auth, deployment). Use when working with collections, live queries, optimistic updates, forms, validation, routing, URL parameters, navigation, server functions, or full-stack TanStack Start apps.
allowed-tools: Read, Grep, Glob
---

# TanStack Developer Guide

This skill provides comprehensive patterns and best practices for the TanStack ecosystem in React applications:

- **TanStack Query/DB**: Data fetching, caching, collections, live queries, and optimistic updates
- **TanStack Form**: Form state management, validation, and field components
- **TanStack Router**: File-based routing, type-safe navigation, and URL parameters

## Quick Start

For detailed examples and patterns, refer to the following files in the `references/` directory:

- `references/query-patterns.md` - TanStack Query and TanStack DB patterns
- `references/form-patterns.md` - TanStack Form patterns and components
- `references/router-patterns.md` - TanStack Router patterns and navigation

Detailed per-rule best practices (consolidated from the former standalone skills):

- `references/query/` - TanStack Query best-practice rules (query keys, caching, mutations, prefetching, SSR)
- `references/router/` - TanStack Router best-practice rules (type safety, loaders, search params, navigation)
- `references/start/` - TanStack Start best-practice rules (server functions, middleware, auth, SSR, deployment)

---

## TanStack Query/DB Overview

TanStack DB extends TanStack Query with collections, live queries, and optimistic mutations. Key principle: load data into typed collections and consume through live queries that auto-update on data changes.

### Critical Rules

1. **Never Use React Query Patterns with Collections** - Collections have built-in mutation handling. Do NOT use `useMutation` + `invalidateQueries`.

2. **Always Share Collection Instances** - Creating new collection instances for mutations causes "key not found" errors. The data-fetching hook must expose the collection, and mutation hooks must receive it as a parameter.

3. **Configure Persistence Handlers** - Put server writes in collection handlers (`onInsert`, `onUpdate`, `onDelete`), not mutation hooks.

4. **Single Canonical Collection Pattern** - Create ONE collection per entity type. Use live queries for filtered views.

5. **Check Field Changes Properly** - Verify fields actually changed in `onUpdate`, not just that they exist.

### Basic Collection Setup

```typescript
import { createCollection } from '@tanstack/react-db';
import { queryCollectionOptions } from '@tanstack/query-db-collection';
import { z } from 'zod';

const itemSchema = z.object({
  id: z.string(),
  name: z.string().min(1),
  status: z.enum(['active', 'archived']),
});

const itemCollection = createCollection(
  queryCollectionOptions({
    queryKey: ['items'],
    queryFn: async () => (await fetch('/api/items')).json(),
    queryClient,
    getKey: (item) => item.id,
    schema: itemSchema,
  })
);
```

### Sharing Collection Instance (Critical Pattern)

```typescript
// CORRECT - share the instance
export function useItems() {
  const collection = useMemo(() => createItemsCollection(), []);
  const { data } = useLiveQuery(collection);
  return { data, collection }; // Expose collection
}

export function useUpdateItem(collection: ItemsCollection) {
  return (id, data) => collection.update(id, data);
}
```

---

## TanStack Form Overview

TanStack Form provides headless form logic with automatic type inference and flexible validation.

### Core Principles

- **Type Safety**: Types are inferred from default values - avoid manual generic declarations.
- **Headless Design**: Build UI components to match your design system.
- **Schema-First Validation**: Use Zod for cleaner, more maintainable validation.

### Basic Form Setup with `createFormHook`

```typescript
import { createFormHookContexts, createFormHook } from '@tanstack/react-form'

export const { fieldContext, formContext, useFieldContext } =
  createFormHookContexts()

export const { useAppForm } = createFormHook({
  fieldContext,
  formContext,
  fieldComponents: {
    TextField,
    SelectField,
  },
  formComponents: {
    SubmitButton,
  },
})
```

### Form Initialization

```typescript
const form = useAppForm({
  defaultValues: {
    username: '',
    email: '',
    age: 0,
  },
  validators: {
    onChange: schema,
  },
  onSubmit: async ({ value }) => {
    // Handle submission
  },
})
```

### Async Validation with Debouncing

```typescript
<form.Field
  name="username"
  asyncDebounceMs={500}
  validators={{
    onChangeAsync: async ({ value }) => {
      const isAvailable = await checkUsernameAvailability(value)
      return isAvailable ? undefined : 'Username already taken'
    },
  }}
/>
```

---

## TanStack Router Overview

TanStack Router provides type-safe file-based routing with first-class TypeScript support.

### Core Principles

- **Type-Safe Routing**: Embrace type-safe routing as the primary benefit.
- **File-Based Routes**: Use file-based routing for scalability.
- **Generated Route Tree**: Leverage the generated route tree for type safety.

### File Structure

```
src/routes/
├── __root.tsx          # Root layout with providers
├── _authenticated.tsx  # Auth layout wrapper
├── index.tsx          # Home page (/)
├── posts/
│   ├── index.tsx      # /posts
│   └── $postId.tsx    # /posts/:postId (typed params)
└── settings/
    ├── _layout.tsx    # Settings layout
    └── profile.tsx    # /settings/profile
```

### Basic Route with Search Params

```typescript
import { createFileRoute } from '@tanstack/react-router'
import { z } from 'zod'

const searchSchema = z.object({
  page: z.number().min(1).catch(1),
  search: z.string().optional(),
})

export const Route = createFileRoute('/posts/')({
  validateSearch: searchSchema,
  component: PostsList,
})

function PostsList() {
  const { page, search } = Route.useSearch()
  // Use search params...
}
```

### Authentication Layout

```typescript
// routes/_authenticated.tsx
import { createFileRoute, redirect, Outlet } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async ({ location }) => {
    const isAuthenticated = checkAuth()
    if (!isAuthenticated) {
      throw redirect({
        to: '/login',
        search: { redirect: location.href },
      })
    }
  },
  component: () => <Outlet />,
})
```

### Type-Safe Navigation

```typescript
import { Link, useNavigate } from '@tanstack/react-router'

function Navigation() {
  const navigate = useNavigate()

  return (
    <>
      <Link
        to="/posts/$postId"
        params={{ postId: '123' }}
        search={{ tab: 'comments' }}
      >
        View Post
      </Link>

      <button onClick={() => navigate({ to: '/posts', search: { page: 1 } })}>
        Go to Posts
      </button>
    </>
  )
}
```

---

## TanStack Query — Best Practices

Comprehensive guidelines for implementing TanStack Query (React Query) patterns in React applications: query keys, caching, mutations, error handling, prefetching, parallel queries, infinite queries, SSR integration, performance, and offline support. Distinct from the collections/live-query model described in "TanStack Query/DB Overview" above — apply these when using classic `useQuery`/`useMutation` hooks directly (not TanStack DB collections).

### When to Apply

- Creating new data fetching logic
- Setting up query configurations
- Implementing mutations and optimistic updates
- Configuring caching strategies
- Integrating with SSR/SSG
- Refactoring existing data fetching code

### Rule Categories by Priority

| Priority | Category | Rules | Impact |
|----------|----------|-------|--------|
| CRITICAL | Query Keys | 5 rules | Prevents cache bugs and data inconsistencies |
| CRITICAL | Caching | 5 rules | Optimizes performance and data freshness |
| HIGH | Mutations | 6 rules | Ensures data integrity and UI consistency |
| HIGH | Error Handling | 3 rules | Prevents poor user experiences |
| MEDIUM | Prefetching | 4 rules | Improves perceived performance |
| MEDIUM | Parallel Queries | 2 rules | Enables dynamic parallel fetching |
| MEDIUM | Infinite Queries | 3 rules | Prevents pagination bugs |
| MEDIUM | SSR Integration | 4 rules | Enables proper hydration |
| LOW | Performance | 4 rules | Reduces unnecessary re-renders |
| LOW | Offline Support | 2 rules | Enables offline-first patterns |

### Query Keys (Prefix: `qk-`)

- `qk-array-structure` — Always use arrays for query keys
- `qk-include-dependencies` — Include all variables the query depends on
- `qk-hierarchical-organization` — Organize keys hierarchically (entity → id → filters)
- `qk-factory-pattern` — Use query key factories for complex applications
- `qk-serializable` — Ensure all key parts are JSON-serializable

### Caching (Prefix: `cache-`)

- `cache-stale-time` — Set appropriate staleTime based on data volatility
- `cache-gc-time` — Configure gcTime for inactive query retention
- `cache-defaults` — Set sensible defaults at QueryClient level
- `cache-invalidation` — Use targeted invalidation over broad patterns
- `cache-placeholder-vs-initial` — Understand placeholder vs initial data differences

### Mutations (Prefix: `mut-`)

- `mut-invalidate-queries` — Always invalidate related queries after mutations
- `mut-optimistic-updates` — Implement optimistic updates for responsive UI
- `mut-rollback-context` — Provide rollback context from onMutate
- `mut-error-handling` — Handle mutation errors gracefully
- `mut-loading-states` — Use isPending for mutation loading states
- `mut-mutation-state` — Use useMutationState for cross-component tracking

### Error Handling (Prefix: `err-`)

- `err-error-boundaries` — Use error boundaries with useQueryErrorResetBoundary
- `err-retry-config` — Configure retry logic appropriately
- `err-fallback-data` — Provide fallback data when appropriate

### Prefetching (Prefix: `pf-`)

- `pf-intent-prefetch` — Prefetch on user intent (hover, focus)
- `pf-route-prefetch` — Prefetch data during route transitions
- `pf-stale-time-config` — Set staleTime when prefetching
- `pf-ensure-query-data` — Use ensureQueryData for conditional prefetching

### Infinite Queries (Prefix: `inf-`)

- `inf-page-params` — Always provide getNextPageParam
- `inf-loading-guards` — Check isFetchingNextPage before fetching more
- `inf-max-pages` — Consider maxPages for large datasets

### SSR Integration (Prefix: `ssr-`)

- `ssr-dehydration` — Use dehydrate/hydrate pattern for SSR
- `ssr-client-per-request` — Create QueryClient per request
- `ssr-stale-time-server` — Set higher staleTime on server
- `ssr-hydration-boundary` — Wrap with HydrationBoundary

### Parallel Queries (Prefix: `parallel-`)

- `parallel-use-queries` — Use useQueries for dynamic parallel queries
- `query-cancellation` — Implement query cancellation properly

### Performance (Prefix: `perf-`)

- `perf-select-transform` — Use select to transform/filter data
- `perf-structural-sharing` — Leverage structural sharing
- `perf-notify-change-props` — Limit re-renders with notifyOnChangeProps
- `perf-placeholder-data` — Use placeholderData for instant UI

### Offline Support (Prefix: `offline-`)

- `network-mode` — Configure network mode for offline support
- `persist-queries` — Configure query persistence for offline support

---

## TanStack Router — Best Practices

Comprehensive guidelines for implementing TanStack Router patterns in React applications: type safety, route organization, data loading, search params, error handling, navigation, code splitting, preloading, and route context.

### When to Apply

- Setting up application routing
- Creating new routes and layouts
- Implementing search parameter handling
- Configuring data loaders
- Setting up code splitting
- Integrating with TanStack Query
- Refactoring navigation patterns

### Rule Categories by Priority

| Priority | Category           | Rules   | Impact                                          |
| -------- | ------------------ | ------- | ------------------------------------------------ |
| CRITICAL | Type Safety        | 4 rules | Prevents runtime errors and enables refactoring |
| CRITICAL | Route Organization | 5 rules | Ensures maintainable route structure            |
| HIGH     | Router Config      | 1 rule  | Global router defaults                          |
| HIGH     | Data Loading       | 6 rules | Optimizes data fetching and caching             |
| HIGH     | Search Params      | 5 rules | Enables type-safe URL state                     |
| HIGH     | Error Handling     | 1 rule  | Handles 404 and errors gracefully               |
| MEDIUM   | Navigation         | 5 rules | Improves UX and accessibility                   |
| MEDIUM   | Code Splitting     | 3 rules | Reduces bundle size                             |
| MEDIUM   | Preloading         | 3 rules | Improves perceived performance                  |
| LOW      | Route Context      | 3 rules | Enables dependency injection                    |

### Type Safety (Prefix: `ts-`)

- `ts-register-router` — Register router type for global inference
- `ts-use-from-param` — Use `from` parameter for type narrowing
- `ts-route-context-typing` — Type route context with createRootRouteWithContext
- `ts-query-options-loader` — Use queryOptions in loaders for type inference

### Router Config (Prefix: `router-`)

- `router-default-options` — Configure router defaults (scrollRestoration, defaultErrorComponent, etc.)

### Route Organization (Prefix: `org-`)

- `org-file-based-routing` — Prefer file-based routing for conventions
- `org-route-tree-structure` — Follow hierarchical route tree patterns
- `org-pathless-layouts` — Use pathless routes for shared layouts
- `org-index-routes` — Understand index vs layout routes
- `org-virtual-routes` — Understand virtual file routes

### Data Loading (Prefix: `load-`)

- `load-use-loaders` — Use route loaders for data fetching
- `load-loader-deps` — Define loaderDeps for cache control
- `load-ensure-query-data` — Use ensureQueryData with TanStack Query
- `load-deferred-data` — Split critical and non-critical data
- `load-error-handling` — Handle loader errors appropriately
- `load-parallel` — Leverage parallel route loading

### Search Params (Prefix: `search-`)

- `search-validation` — Always validate search params
- `search-type-inheritance` — Leverage parent search param types
- `search-middleware` — Use search param middleware
- `search-defaults` — Provide sensible defaults
- `search-custom-serializer` — Configure custom search param serializers

### Error Handling (Prefix: `err-`)

- `err-not-found` — Handle not-found routes properly

### Navigation (Prefix: `nav-`)

- `nav-link-component` — Prefer Link component for navigation
- `nav-active-states` — Configure active link states
- `nav-use-navigate` — Use useNavigate for programmatic navigation
- `nav-relative-paths` — Understand relative path navigation
- `nav-route-masks` — Use route masks for modal URLs

### Code Splitting (Prefix: `split-`)

- `split-lazy-routes` — Use .lazy.tsx for code splitting
- `split-critical-path` — Keep critical config in main route file
- `split-auto-splitting` — Enable autoCodeSplitting when possible

### Preloading (Prefix: `preload-`)

- `preload-intent` — Enable intent-based preloading
- `preload-stale-time` — Configure preload stale time
- `preload-manual` — Use manual preloading strategically

### Route Context (Prefix: `ctx-`)

- `ctx-root-context` — Define context at root route
- `ctx-before-load` — Extend context in beforeLoad
- `ctx-dependency-injection` — Use context for dependency injection

---

## TanStack Start — Best Practices

Comprehensive guidelines for implementing TanStack Start patterns in full-stack React applications: server functions, middleware, SSR, authentication, and deployment.

### When to Apply

- Creating server functions for data mutations
- Setting up middleware for auth/logging
- Configuring SSR and hydration
- Implementing authentication flows
- Handling errors across client/server boundary
- Organizing full-stack code
- Deploying to various platforms

### Rule Categories by Priority

| Priority | Category          | Rules   | Impact                      |
| -------- | ----------------- | ------- | ---------------------------- |
| CRITICAL | Server Functions  | 5 rules | Core data mutation patterns |
| CRITICAL | Security          | 4 rules | Prevents vulnerabilities    |
| HIGH     | Middleware        | 4 rules | Request/response handling   |
| HIGH     | Authentication    | 4 rules | Secure user sessions        |
| MEDIUM   | API Routes        | 1 rule  | External endpoint patterns  |
| MEDIUM   | SSR               | 6 rules | Server rendering patterns   |
| MEDIUM   | Error Handling    | 3 rules | Graceful failure handling   |
| MEDIUM   | Environment       | 1 rule  | Configuration management    |
| LOW      | File Organization | 3 rules | Maintainable code structure |
| LOW      | Deployment        | 2 rules | Production readiness        |

### Server Functions (Prefix: `sf-`)

- `sf-create-server-fn` — Use createServerFn for server-side logic
- `sf-input-validation` — Always validate server function inputs
- `sf-method-selection` — Choose appropriate HTTP method
- `sf-error-handling` — Handle errors in server functions
- `sf-response-headers` — Customize response headers when needed

### Security (Prefix: `sec-`)

- `sec-validate-inputs` — Validate all user inputs with schemas
- `sec-auth-middleware` — Protect routes with auth middleware
- `sec-sensitive-data` — Keep secrets server-side only
- `sec-csrf-protection` — Implement CSRF protection for mutations

### Middleware (Prefix: `mw-`)

- `mw-request-middleware` — Use request middleware for cross-cutting concerns
- `mw-function-middleware` — Use function middleware for server functions
- `mw-context-flow` — Properly pass context through middleware
- `mw-composability` — Compose middleware effectively

### Authentication (Prefix: `auth-`)

- `auth-session-management` — Implement secure session handling
- `auth-route-protection` — Protect routes with beforeLoad
- `auth-server-functions` — Verify auth in server functions
- `auth-cookie-security` — Configure secure cookie settings

### API Routes (Prefix: `api-`)

- `api-routes` — Create API routes for external consumers

### SSR (Prefix: `ssr-`)

- `ssr-data-loading` — Load data appropriately for SSR
- `ssr-hydration-safety` — Prevent hydration mismatches
- `ssr-streaming` — Implement streaming SSR for faster TTFB
- `ssr-selective` — Apply selective SSR when beneficial
- `ssr-prerender` — Configure static prerendering and ISR

### Environment (Prefix: `env-`)

- `env-functions` — Use environment functions for configuration

### Error Handling (Prefix: `err-`)

- `err-server-errors` — Handle server function errors
- `err-redirects` — Use redirects appropriately
- `err-not-found` — Handle not-found scenarios

### File Organization (Prefix: `file-`)

- `file-separation` — Separate server and client code
- `file-functions-file` — Use .functions.ts pattern
- `file-shared-validation` — Share validation schemas

### Deployment (Prefix: `deploy-`)

- `deploy-env-config` — Configure environment variables
- `deploy-adapters` — Choose appropriate deployment adapter

---

## Validation Checklist

Before finishing a task involving TanStack:

### Query/DB
- [ ] Collection instances are shared between data-fetching and mutation hooks
- [ ] Persistence handlers (`onInsert`, `onUpdate`, `onDelete`) are configured
- [ ] No `useMutation` + `invalidateQueries` patterns with collections
- [ ] One canonical collection per entity type
- [ ] Field changes properly verified in `onUpdate` handlers

### Form
- [ ] Use `createFormHook` with `useAppForm` instead of raw `useForm` for consistency
- [ ] Provide complete default values for proper type inference
- [ ] Use Zod schemas for validation when possible
- [ ] Debounce async validations (minimum 500ms recommended)
- [ ] Prevent default on form submission
- [ ] Display errors with proper accessibility (`role="alert"`)

### Router
- [ ] Route path in `createFileRoute` matches file location
- [ ] Search params use Zod validation with proper defaults (`.catch()`)
- [ ] Loader dependencies are correctly specified in `loaderDeps`
- [ ] Authentication routes use `beforeLoad` with proper redirects
- [ ] Navigation uses typed `Link` or `useNavigate` hooks
- [ ] Error boundaries are implemented at route level

### General
- [ ] Run `pnpm run typecheck` and `pnpm run test`

---

For complete examples, edge cases, and advanced patterns, see the reference files in this skill directory.
