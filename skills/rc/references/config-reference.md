# Configuration Reference — removed

RC no longer reads a `.rc/config.toml`. That file configured the removed RC CLI binary
(`ide`, `model`, `reasoning_effort`, `access_mode`, `timeout`, `[tasks.run]`, `[fix_reviews]`,
`[fetch_reviews]`, `[exec]`, …); with the CLI gone, nothing consumes it.

## Where behavior is configured now

- **Model & reasoning per role** — the `model:` and `effort:` frontmatter of each agent
  (`agents/*.md`) and skill (`skills/*/SKILL.md`).
- **Hook behavior** — env vars read by `hooks/scripts/_lib.sh`: `RC_HOOK_PROFILE`
  (`minimal` | `standard` | `strict`, default `standard`) and `RC_DISABLED_HOOKS`
  (comma-separated hook names). `RC_INSTINCTS=0` disables the `observe` hook.
- **Everything else** — decided per-invocation inside the relevant skill/command, not a config file.

There is no plugin-native config file today. If one is introduced later, document it here.
