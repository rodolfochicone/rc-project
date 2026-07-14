# Interactive Run Input — Task List

## Tasks

| # | Title | Status | Complexity | Dependencies |
|---|-------|--------|------------|--------------|
| 01 | Input coordinator model types and runtime config flags | completed | low | — |
| 02 | Awaiting-input event kind and payload | completed | low | — |
| 03 | Concrete input coordinator (channel mailbox) | completed | medium | task_01 |
| 04 | Interactive ACP permission callback and live-session continue | completed | high | task_01, task_02 |
| 05 | Interactive exec turn loop | completed | high | task_02, task_03, task_04 |
| 06 | Run manager wiring, SendInput, and snapshot pending_input | completed | high | task_01, task_03, task_05 |
| 07 | HTTP input endpoint and OpenAPI surface | completed | medium | task_06 |
| 08 | Frontend input data layer | completed | medium | task_07 |
| 09 | Frontend response panel and option parsing | completed | high | task_08 |
