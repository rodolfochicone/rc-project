# Shareable Project Memory — Task List

## Tasks

| # | Title | Status | Complexity | Dependencies |
|---|-------|--------|------------|--------------|
| 01 | Mirror format: marshal, parse, and filename for memory files | completed | medium | — |
| 02 | Store.Import: transactional most-recent-wins batch upsert | completed | medium | — |
| 03 | `rc memory export` subcommand (DB → mirror files) | completed | low | task_01 |
| 04 | `rc memory import` subcommand (mirror files → DB) + integration tests | completed | medium | task_01, task_02, task_03 |
| 05 | Docs: package tenet update and rc-project-memory skill | completed | low | task_03, task_04 |
