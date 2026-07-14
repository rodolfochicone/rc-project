# LESSONS — auto-maintained by scripts/lessons.mjs

> Machine-owned. Do NOT hand-edit. Changes are overwritten on the next `lessons.mjs` write.
> Canonical state lives in `.rc/lessons.json`. Edit lessons only via the script.
> promote_threshold=2 distinct features · window_days=45 · quarantine_threshold=2

## Confirmed (load these at plan/design time)

Corroborated across multiple features. Safe to apply as guidance.

_none_

## Candidates (under observation — do NOT load as guidance yet)

Seen once or not yet corroborated. Tracked, not trusted.

### L-001 — Link pendurado tem duas causas raiz opostas: se o arquivo existe fora do lugar, mova o arquivo (o link estava certo); so apague a referencia quando o conteudo nao existe em lugar nenhum — e ai apague a promessa inteira (secao, bullet de trigger, claim na description), nao so o link clicavel.
- signal: `gate_fail` · recurrence: 1 feature(s) · harmful: 0
- features: asset-integrity
- evidence: scripts/plugin-smoke.mjs dangling-asset; skills/rc-autoresearch/SKILL.md vs skills/rc-bubbletea/SKILL.md
- last seen: 2026-07-14T16:12:52Z

### L-002 — Quando o nome de um arquivo shipado e a referencia da skill divergem, o conteudo do proprio arquivo decide quem esta stale; depois de renomear um script, confira o modo contra os irmaos (644 vs 755).
- signal: `gate_fail` · recurrence: 1 feature(s) · harmful: 0
- features: asset-integrity
- evidence: skills/rc-smux-rc-pairing/scripts/run-compozy-start.sh (banner imprimia 'Usage: run-rc-start.sh')
- last seen: 2026-07-14T16:12:52Z

## Quarantined (failed when applied — ignore)

A confirmed lesson that recurred alongside failure. Kept for the maintainer to review.

_none_
