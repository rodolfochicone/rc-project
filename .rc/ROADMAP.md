# Roadmap — Integridade de assets das skills

> Autonomous loop source of truth. `rc-loop` reads this file each iteration (phase goal +
> sub-items). The per-phase tasks are derived just-in-time by rc-create-tasks against the
> codebase, workflow memory, and confirmed lessons. A checkbox flips to `[x]` only when a
> verification PASS is recorded for that phase.

**Legend:** `[ ]` not started · `[~]` in progress · `[x]` done (verification PASS)

**Hard dependency order:** nenhuma dependência entre as fases (arquivos disjuntos); rodar na
ordem 1 → 2. Never run phases in parallel.

**Origem do backlog:** os achados do `plugin-smoke` (check `dangling asset`, v2.4.0+). Não foram
inventados — cada um é um link que uma skill publica hoje e que não resolve. O gate é o oráculo:
`node scripts/plugin-smoke.mjs` sai 0.

---

## Phase 1 — Links quebrados nas skills de referência `[x]`

> Done when: `node scripts/plugin-smoke.mjs` não reporta nenhum `dangling asset` em
> `rc-autoresearch`, `rc-bubbletea` ou `rc-zod`.

- [x] `rc-autoresearch` — o `SKILL.md` linka `references/eval-guide.md`, mas o arquivo está em
      `skills/rc-autoresearch/eval-guide.md` e a skill não tem pasta `references/`
- [x] `rc-bubbletea` — o `SKILL.md` manda ler `references/effects.md`, que não existe
- [x] `rc-zod` — `SKILL.md` e `AGENTS.md` linkam `references/_sections.md` e
      `assets/templates/_template.md`; nenhum dos dois existe

## Phase 2 — Resquícios de template e do fork Compozy `[x]`

> Done when: `node scripts/plugin-smoke.mjs` sai 0 — gate inteiro verde, zero problemas.

- [x] `rc-skill-best-practices` — o `SKILL.md` manda usar `assets/skill-template.md`; o arquivo
      real é `assets/SKILL.template.md`
- [x] `rc-smux-rc-pairing` — o `SKILL.md` manda executar `scripts/run-rc-start.sh`; o script que
      existe é `run-compozy-start.sh` (resquício do fork Compozy)
